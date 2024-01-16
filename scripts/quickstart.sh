#!/usr/bin/env bash

DEBUG_FLAG="false"

log() {
  if [ "$#" -ge 2 ]; then
    type="$1"
    shift
  else
    type="debug"
  fi
  message="$*"
  color_start=""

  # check if color is supported (should be > 256 within xterm)
  if [ "$(tput colors)" -gt 0 ]; then
    case "$type" in
      info)
        color_code=14
        ;;
      warn)
        color_code=11
        ;;
      error)
        color_code=9
        ;;
      *)
        color_code=15
        ;;
    esac
    color_start="$(tput setaf "$color_code")"
  fi

  printf "%s%s: %s\n" "$color_start" "$type" "$message"
  if [ -n "$color_start" ]; then
    tput init
  fi
}

info() {
  log "info" "$@"
}

error() {
  log "error" "$@"
}

warn() {
  log "warn" "$@"
}

_commands=(
  "docker"
  "git"
)

for cmd in "${_commands[@]}"; do
  if ! command -v "$cmd" &> /dev/null; then
    error "required command '$cmd' not found"
    exit 1
  fi
done

repo_path="$(git rev-parse --show-toplevel)"
impl_path="$repo_path/impl"

if ! pushd "$impl_path" > /dev/null; then
  error "unable to find path '$impl_path'"
  exit 1
fi

docker_path="$impl_path/build/Dockerfile"
if [ ! -f "$docker_path" ]; then
  error "unable to find build file '$docker_path'"
  exit 1
fi

# defaults ---------------------------------------------------------------------
# docker build
opt_commit_hash="$(git rev-parse HEAD)"
opt_tag="did-dht:latest"

# docker run
opt_detach=""
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
      echo "  -d, --detach        run the container in the background (default: false)"
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
    -d | --detach)
      opt_detach="--detach"
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
      warn "skipping unknown option '$1'"
      opt_erroneous_option="true"
      shift
      ;;
  esac
done

if [ "$opt_erroneous_option" == "true" ]; then
  echo ""
  warn "one or more options were skipped"
  warn "  run \`$0 --help\` for more information"
  exit 1
fi

if git cat-file -e "$opt_commit_hash"; then
  info "building from commit $opt_commit_hash"
else
  error "please specify a valid commit hash, $opt_commit_hash is does not exist"
  exit 1
fi

if [ "$DEBUG_FLAG" == "true" ]; then
  info "running in debug mode"
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
  info "skipping run"
  exit 0
fi

if docker ps --all --format '{{.Names}}' | grep --quiet "$opt_name"; then
  error "container $opt_name already exists"
  exit 1
fi

if [ -z "$opt_detach" ]; then
  info "running in foreground"
  info "use ctrl-p ctrl-q to detach from the container (send to background)"
  echo ""
fi

$_DEBUG docker run \
  $opt_detach \
  --interactive \
  --tty \
  --publish "$opt_port" \
  --name "$opt_name" \
  "$opt_remove" \
  "$opt_tag"
