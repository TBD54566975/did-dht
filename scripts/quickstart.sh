#!/usr/bin/env bash

_commands=(
  "docker"
  "git"
)

for cmd in "${_commands[@]}"; do
  if ! command -v "$cmd" &> /dev/null; then
    echo "error: $cmd could not be found"
    exit 1
  fi
done

repo_path="$(git rev-parse --show-toplevel)"
impl_path="$repo_path/impl"

if ! pushd "$impl_path" > /dev/null; then
  echo "error: unable to find '$impl_path'"
  exit 1
fi

docker_path="$impl_path/build/Dockerfile"
if [ ! -f "$docker_path" ]; then
  echo "error: unable to find Dockerfile '$docker_path'"
  exit 1
fi

# defaults ---------------------------------------------------------------------
# docker build
opt_commit_hash="$(git rev-parse HEAD)"
opt_tag="did-dht:latest"

# docker run
opt_detach="--detach"
opt_remove="--rm"
opt_name="did-dht-server"
opt_port="8305:8305"
opt_skip_run="false"
opt_erroneous_option="false"

for opt in "$@"; do
  case "$opt" in
    -h | --help)
      echo "Usage: $0 [options]"
      echo ""
      echo "Builds and runs the did-dht server"
      echo ""
      echo "Options"
      echo "  -h, --help          show this help message and exit"
      echo "  -c, --commit=<hash> commit hash for \`docker build\` (default: HEAD)"
      echo "  -t, --tag=<tag>     tag name for \`docker build\` (default: did-dht:latest)"
      echo "  -a, --attach        run the container in the foreground (default: true)"
      echo "  -k, --keep          keep the container after it exits (default: false)"
      echo "  -n, --name=<name>   name to give the container (default: did-dht-server)"
      echo "  -p, --port=<port>   ports to publish the host/container (default: 8305:8305)"
      echo "  --skip-run          skip running the container (default: false)"
      exit 0
      ;;

    # build options ------------------------------------------------------------
    -c=* | --commit=*)
      opt_commit_hash="${opt#*=}"
      shift
      ;;

    -t=* | --tag=*)
      opt_tag="${opt#*=}"
      shift
      ;;

    # run options --------------------------------------------------------------
    -a | --attach)
      unset opt_detach
      shift
      ;;

    -k | --keep)
      unset opt_remove
      shift
      ;;

    -n=* | --name=*)
      opt_name="${opt#*=}"
      shift
      ;;

    -p=* | --port=*)
      opt_port="${opt#*=}"
      shift
      ;;

    --skip-run)
      opt_skip_run="true"
      shift
      ;;

    *)
      echo "warn: skipping unknown option '$1'"
      opt_erroneous_option="true"
      shift
      ;;
  esac
done

if [ "$opt_erroneous_option" == "true" ]; then
  echo ""
  echo "warning: one or more options were skipped"
  echo "         run \`$0 --help\` for more information"
  echo ""
  exit 1
fi

if git cat-file -e "$opt_commit_hash"; then
  echo "info: building from commit $opt_commit_hash"
else
  echo "error: commit $opt_commit_hash does not exist"
  exit 1
fi

DEBUG_FLAG="false"
if [ "$DEBUG_FLAG" == "true" ]; then
  echo "info: running in debug mode"
  _DEBUG="echo"
else
  unset _DEBUG
fi

echo "created image $opt_tag at commit ${opt_commit_hash:0:8}"
$_DEBUG docker build \
  --build-arg GIT_COMMIT_HASH="$opt_commit_hash" \
  --tag "$opt_tag" \
  --file "$docker_path" \
  "$impl_path"

echo ""
if [ "$opt_skip_run" == "true" ]; then
  echo "info: skipping run"
  exit 0
fi

if docker ps --all --format '{{.Names}}' | grep --quiet "$opt_name"; then
  echo "error: container $opt_name already exists"
  exit 1
fi

printf "info: running container: "
$_DEBUG docker run \
  "$opt_detach" \
  "$opt_remove" \
  --interactive \
  --tty \
  --publish "$opt_port" \
  --name "$opt_name" \
  "$opt_tag"
