DIRS="api
api/proto
cmd
config
internal
internal/app
internal/app/core
internal/pkg
scripts"

for dir in $DIRS; do
	mkdir -p $dir

done
