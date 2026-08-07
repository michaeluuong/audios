package mp3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/muesli/gomusicbrainz"
	_ "golang.org/x/image/webp"

	"github.com/michaeluuong/utilize/filing"
	"github.com/michaeluuong/utilize/reflections"
	"github.com/michaeluuong/utilize/stringy"
)

const (
	CoverArtArchiveUrl = "https://coverartarchive.org"
	email              = "qualude@aol.com"
)

const (
	slogLevelCover = slogLevelNoOp
)

type CAAResponse struct {
	Images []struct {
		IApproved  bool   `json:"approved"`
		IsBack     bool   `json:"back"`
		Comment    string `json:"comment"`
		Edit       int    `json:"edit"`
		IsFront    bool   `json:"front"`
		Id         int    `json:"Id"`
		Image      string `json:"image"`
		Thumbnails struct {
			T250  string `json:"250"`
			T500  string `json:"500"`
			T1200 string `json:"1200"`
			Small string `json:"small"`
			Large string `json:"large"`
		} `json:"thumbnails"`
		Types []string `json:"types"`
	} `json:"images"`
	Release string `json:"release"`
}

/*
curl -v -L -o mw https://archive.org/download/mbid-69ea56cf-822c-4036-b1a5-8072831078bc/index.json
ls -alt
view mw
curl https://coverartarchive.org/release/69ea56cf-822c-4036-b1a5-8072831078bc/44504132347.jpg
curl -v -L https://coverartarchive.org/release/69ea56cf-822c-4036-b1a5-8072831078bc/44504132347.jpg
ls -alt
curl -v -L -o mw.jpg https://coverartarchive.org/release/69ea56cf-822c-4036-b1a5-8072831078bc/44504132347.jpg
*/

// MBIDRelease searches for musicbrainz releases by artist and album.
//   - artist the name of the artist to search for (this will also be used for creditname to reduce the search results)
//   - album the name of the album to search for
//
// NOTE: since musicbrainz rate limits to 1 request per second this function will sleep for 1 second.
//
// Return all musicbrainz releases that matched the search critera or an error
//   - if either artist or album were not provided
//   - the search terms did not match any releases
func MBIDRelease(artist, album string) ([]gomusicbrainz.Release, error) {
	slog.Log(context.TODO(), slogLevelCover, "started", "artist", artist, "album", album)
	if artist == "" || album == "" {
		return nil, errors.New("artist(" + artist + ") and album(" + album + ") are required")

	}

	// Create client and set mandatory identification
	client := gomusicbrainz.NewWS2Client()
	client.SetClientInfo("audiostag", "1.0.0", email)

	// Search for a specific release
	// Query: release name, artist name and creditname (which will have the same value as artist to reduce results)
	searchTerm := fmt.Sprintf("release:\"%s\" AND artist:\"%s\" AND creditname:\"%s\"",
		regexp.QuoteMeta(album), regexp.QuoteMeta(artist), regexp.QuoteMeta(artist))
	slog.Debug("search term for release", "searchTerm", searchTerm)
	resp, err := client.SearchRelease(searchTerm, -1, -1)
	//t := *gomusicbrainz.ReleaseResponse
	if err != nil {
		slog.Error("SearchRelease()", "err", err)
		return nil, err

	}

	// MusicBrainz limits anonymous requests to 1 request per second
	if taggerConfig.CoverSleepTime > 0 {
		time.Sleep(time.Duration(taggerConfig.CoverSleepTime) * time.Second)

	}

	slog.Log(context.TODO(), slogLevelCover, "finished", "resp.Releases", resp.Releases)
	return resp.Releases, nil

}

// RequestCAAResponse attempts to GET cover art URL information from coverartarchive.org.
//   - mbid is the musicbrainz id for the album
//
// Return the response from coverartarchiveorg or and error
//   - if unable to acquire a new http request
//   - if http response was anything other than 200 and/or the request returned an error
//   - if unable to unmarshal the responss to the CAAResponse object or there are no images in the response
func RequestCAAResponse(mbid string) (*CAAResponse, error) {
	slog.Log(context.TODO(), slogLevelCover, "started", "mbid", mbid)
	url := fmt.Sprintf("%s/release/%s", CoverArtArchiveUrl, mbid)

	req, err := http.NewRequest("GET", url, nil)
	fmt.Printf("req: %v\n", req)
	if err != nil {
		slog.Error("http.NewRequest()|could not create GET request for cover art", "mbid", mbid, "url", url, "err", err)

	}
	req.Header.Set("User-Agent", "audiostag/1.0 ( qualude@aol.com )")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		slog.Error("Client.Do()", "mbid", mbid, "req.Header", req.Header, "resp", resp, "err", err)
		return nil, fmt.Errorf("no cover art found for mbid %s resp.StatusCode %d: %w", mbid, resp.StatusCode, err)

	}
	defer resp.Body.Close()

	var caaResponse CAAResponse
	if err := json.NewDecoder(resp.Body).Decode(&caaResponse); err != nil || len(caaResponse.Images) == 0 {
		slog.Error("json.NewDecoder().Decode()", "mbid", mbid, "err", err)
		return nil, fmt.Errorf("no images found: %w", err)

	}

	slog.Log(context.TODO(), slogLevelCover, "finished", "caaResponse", caaResponse)
	return &caaResponse, nil

}

// findCAAImage tries to find the largest image from teh CAAResponse.
//   - Image
//   - Thumbnails.Large
//   - Thumbnails.T1200
//   - Thumbnails.T500
//   - Thumbnails.Small
//   - Thumbnails.T250
func findCAAImage(caaResponse *CAAResponse) string {
	slog.Log(context.TODO(), slogLevelCover, "started", "caaResponse", caaResponse)
	var imageURL string
	for _, image := range caaResponse.Images {
		if taggerConfig.CoverSize != "" {
			coverSize := strings.ToLower(taggerConfig.CoverSize)
			slog.Debug("CAA_CHECK", "image.Thumbnails", image.Thumbnails)
			// 250, 500, 1200, small, large, front
			if coverSize == "front" {
				imageURL = image.Image

			} else if value, err := reflections.ReflectFieldByName(image.Thumbnails, taggerConfig.CoverSize); err == nil {
				imageURL = value.String()
				slog.Debug("CAA_CHECK", "imageURL", imageURL)

			}

		}

		if imageURL == "" {
			if image.Thumbnails.T1200 != "" {
				imageURL = image.Thumbnails.T1200

			} else {
				if image.IsFront {
					imageURL = image.Image

				} else if image.Thumbnails.Large != "" {
					imageURL = image.Thumbnails.Large

				} else if image.Thumbnails.T1200 != "" {
					imageURL = image.Thumbnails.T1200

				} else if image.Thumbnails.T500 != "" {
					imageURL = image.Thumbnails.T500

				} else if image.Thumbnails.Small != "" {
					imageURL = image.Thumbnails.Small

				} else if image.Thumbnails.T250 != "" {
					imageURL = image.Thumbnails.T250

				}

			}

		}

		if imageURL != "" {
			break

		}

	}

	slog.Log(context.TODO(), slogLevelCover, "finished", "imageURL", imageURL)
	return imageURL

}

// CoverExists tries to find an existing cover art file (e.g. cover.jpg, cover.png, folder.jpg, folder.png).
func CoverExists(dir string) bool {
	slog.Log(context.TODO(), slogLevelCover, "started", "dir", dir)
	filenames := filing.LsEntryName(dir, taggerConfig.ArtworkRegex())
	isCoverExists := false
	if len(filenames) > 0 {
		isCoverExists = true

	}

	slog.Log(context.TODO(), slogLevelCover, "finished", "isCoverExists", isCoverExists)
	return isCoverExists

}

// CoverArt retrieves a cover art image if a cover file does not already exist.
// Attempts to find a cover art file from coverartarchive.org by: retrieving MBID from musicbrainz.org,
// getting release information from coverartarchive.org (with MBID) if any exist,
// pullng the largest cover art image from the response options.
// artist and album are considered to be matches to coverartarchive.org's artist and release fields if they fold into each other or
// the Levenshtien similariy score between them is moderately strict.
//   - artist the name of the artist to search for (the artist is also used for the creditname musicbrainz release field)
//   - album the name of the album to search for
//   - dir is the directory to place the cover art
//   - coverFilenameNoExtOpt is the name of the cover art file (default is "cover")
//
// Return error
//   - if an MBID does not exist for the artist and album
//   - if no release information exists fro the MBID
//     -
func CoverArt(artist, album, dir string, coverFilenameNoExtOpt ...string) (string, error) {
	slog.Log(context.TODO(), slogLevelCover, "started", "artist", artist, "album", album, "dir", dir, "coverFilenameNoExtOpt", coverFilenameNoExtOpt)
	if CoverExists(dir) {
		slog.Debug("CoverExists()|cover art already exists", "dir", dir)
		return "", nil

	}

	var coverFilenameNoExt = "cover"
	if len(coverFilenameNoExtOpt) > 0 && coverFilenameNoExtOpt[0] != "" {
		coverFilenameNoExt = coverFilenameNoExtOpt[0]

	}

	if artist == "" || album == "" {
		return "", errors.New("artist(" + artist + ") and album(" + album + ") are required")

	}
	artistLow, albumLow := strings.ToLower(artist), strings.ToLower(album)
	slog.Info("searching for cover art", "artist", artist, "album", album)

	mbidReleases, err := MBIDRelease(artist, album)
	if len(mbidReleases) == 0 {
		slog.Debug("MBIDRelease()|could not find MBID", "artist", artist, "album", album)
		return "", nil

	} else if err != nil {
		slog.Error("MBID()", "artist", artist, "album", album, "err", err)
		return "", err

	}

	var filename string
	for rNum, release := range mbidReleases {
		rArtist, rAlbum := release.ArtistCredit.NameCredit.Artist.Name, release.Title
		rArtistLow, rAlbumLow := strings.ToLower(rArtist), strings.ToLower(rAlbum)

		if (strings.EqualFold(artist, rArtist) && strings.EqualFold(album, rAlbum)) ||
			(stringy.IsModeratelyStrict(artistLow, rArtistLow) && stringy.IsModeratelyStrict(albumLow, rAlbumLow)) {
			//mbid := release.ReleaseGroup.ID
			mbid := release.ID
			slog.Info("found release", "artist", artist, "album", album, "mbid", mbid)

			caaResponse, err := RequestCAAResponse(mbid)
			slog.Log(context.TODO(), slogLevelNoOp+4, "RequestCAAResponse()", "mbid", mbid, "caaResponse", caaResponse)
			if err != nil {
				slog.Error("RequestCAAResonse()|did not get a CAA response", "artist", artist, "album", album, "mbid", mbid, "err", err)
				//return "", err
				continue

			}

			imageURL := findCAAImage(caaResponse)

			if imageURL != "" {
				filename, err = DownloadPicture(imageURL, dir, coverFilenameNoExt)
				if err != nil {
					slog.Error("DownloadPicture()|could not pull cover art from CAA", "dir", dir, "imageURL", imageURL, "err", err)
					return "", err

				}
				slog.Info("found cover image", "imageURL", imageURL)

				break

			} else {
				slog.Debug("RequestCAAResponse()|could not get image from response", "caaResponse.Images", caaResponse.Images)

			}

		} else {
			slog.Info("release info does not match", "rNum", rNum, "artist", artist, "album", album, "release", release)

		}

	}

	slog.Log(context.TODO(), slogLevelCover, "finished", "filename", filename)
	return filename, nil

}

func DownloadPicture(imageURL, dest, coverFilenameNoExt string) (string, error) {
	slog.Log(context.TODO(), slogLevelCover, "started", "imageURL", imageURL, "dest", dest, "coverFilenameNoExt", coverFilenameNoExt)

	// 1. Get the image
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", err

	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("http.Get()", "StatusCode", resp.StatusCode)
		return "", errors.New("unable to GET image: " + strconv.Itoa(resp.StatusCode))

	}

	// Decode the response into an image
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		if err != nil {
			slog.Error("image.Decode()|could not decode image", "imageURL", imageURL, "coverFilenameNoExt", coverFilenameNoExt, "err", err)
			return "", err

		}

	}

	filename := filepath.Join(dest, coverFilenameNoExt+".jpg")
	fmt.Printf("filename: %s\n", filename)

	// Create the local file
	outfile, err := os.Create(filename)
	if err != nil {
		return "", err

	}
	defer outfile.Close()

	// Write the image data to a jpeg file
	err = jpeg.Encode(outfile, img, &jpeg.Options{Quality: 100})
	if err != nil {
		slog.Error("jpeg.Encode()|could not encode picture", "imageURL", imageURL, "outfile", outfile, "err", err)
		return "", err

	}

	// 5. Stream the response body to the file
	/*_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println("Error saving image:", err)
		slog.Error("io.Copy()", "file", file)
		return "", err

	}*/

	slog.Log(context.TODO(), slogLevelCover, "finished", "filename", filename)
	return filename, nil

}

func ArtworkImage() image.Image {
	return nil

}
