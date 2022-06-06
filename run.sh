#!/bin/sh

up_state_storage() {
  local PORT="$2" # 9502
  local PATH_VOLUME="$3" # $(readlink -f ./ci/dev/data/state_storage)

  docker run --rm -d \
  --name comixifier-state-storage \
  -p "$PORT":6379 \
  -v "$PATH_VOLUME":/data \
  redis redis-server --save 60 1 --loglevel warning
}

down_state_storage() {
  docker stop comixifier-state-storage
}

up_image_storage() {
  local PORT_STORAGE="$2" # 9500
  local PORT_DASHBOARD="$3" # 9501
  local PATH_VOLUME="$4" # $(readlink -f ./ci/dev/data/image_storage)

  docker run --rm -d \
  --name comixifier-image-storage \
  -p "$PORT_DASHBOARD":9000 \
  -p "$PORT_STORAGE":9001 \
  -v "$PATH_VOLUME":/data \
  minio/minio server /data --console-address ":9001"
}

down_image_storage() {
  docker stop comixifier-image-storage
}

up_app() {
  local IMAGE_STORAGE_ENDPOINT="$2" # 127.0.0.1:9501
  local STATE_STORAGE_ENDPOINT="$3" # 127.0.0.1:9502
  local FACE2COMICS_PHONE="$4" # +79111111111
  local FACE2COMICS_APP_ID="$5" # 11111111
  local FACE2COMICS_APP_HASH="$6" # aaaaaaaaaabbbbbbbbbbcccccccccc
  local CUTOUT_API_TOKEN="$7" # aaaaaaaaaabbbbbbbbbbcccccccccc
  local VANCEAI_API_TOKEN="$8" # aaaaaaaaaabbbbbbbbbbcccccccccc

  echo "build..."
  go build -o comixifier .
  echo "start!"
  IMAGE_STORAGE_ENDPOINT="$IMAGE_STORAGE_ENDPOINT" \
  STATE_STORAGE_ENDPOINT="$STATE_STORAGE_ENDPOINT" \
  FACE2COMICS_PHONE="$FACE2COMICS_PHONE" \
  FACE2COMICS_APP_ID="$FACE2COMICS_APP_ID" \
  FACE2COMICS_APP_HASH="$FACE2COMICS_APP_HASH" \
  CUTOUT_API_TOKEN="$CUTOUT_API_TOKEN" \
  VANCEAI_API_TOKEN="$VANCEAI_API_TOKEN" \
  ./comixifier
}

command="$1"
if [ -z "$command" ]
then
 using
 exit 0;
else
 $command $@
fi
