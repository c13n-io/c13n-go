PACKAGE=c13n

# Needed for setting file timestamps to get reproducible archives.
BUILD_DATE="2020-01-01 00:00:00"
BUILD_DATE_STAMP="202001010000.00"

# green prints one line of green text (if the terminal supports it).
function green() {
  echo -e "\e[0;32m${1}\e[0m"
}

# red prints one line of red text (if the terminal supports it).
function red() {
  echo -e "\e[0;31m${1}\e[0m"
}

# reproducible_tar_gzip creates a reproducible tar.gz file of a directory. This
# includes setting all file timestamps and ownership settings uniformly.
function reproducible_tar_gzip() {
  local dir=$1
  local tar_cmd=tar

  # MacOS has a version of BSD tar which doesn't support setting the --mtime
  # flag. We need gnu-tar, or gtar for short to be installed for this script to
  # work properly.
  tar_version=$(tar --version)
  if [[ ! "$tar_version" =~ "GNU tar" ]]; then
    if ! command -v "gtar"; then
      echo "GNU tar is required but cannot be found!"
      echo "On MacOS please run 'brew install gnu-tar' to install gtar."
      exit 1
    fi

    # We have gtar installed, use that instead.
    tar_cmd=gtar
  fi

  # Pin down the timestamp time zone.
  export TZ=UTC

  find "${dir}" -print0 | LC_ALL=C sort -r -z | $tar_cmd \
    "--mtime=${BUILD_DATE}" --no-recursion --null --mode=u+rw,go+r-w,a+X \
    --owner=0 --group=0 --numeric-owner -c -T - | gzip -9n > "${dir}.tar.gz"

  rm -r "${dir}"
}

# reproducible_zip creates a reproducible zip file of a directory. This
# includes setting all file timestamps.
function reproducible_zip() {
  local dir=$1

  # Pin down file name encoding and timestamp time zone.
  export TZ=UTC

  # Set the date of each file in the directory that's about to be packaged to
  # the same timestamp and make sure the same permissions are used everywhere.
  chmod -R 0755 "${dir}"
  touch -t "${BUILD_DATE_STAMP}" "${dir}"
  find "${dir}" -print0 | LC_ALL=C sort -r -z | xargs -0r touch \
    -t "${BUILD_DATE_STAMP}"

  find "${dir}" | LC_ALL=C sort -r | zip -o -X -r -@ "${dir}.zip"

  rm -r "${dir}"
}

# build_release builds the actual release binaries.
#   arguments: <build-system(s)> <ldflags> <package>
function build_release() {
  local sys=$1
  local ldflags=$2
  local package=$3

  green " - Packaging vendor"
  go mod vendor
  reproducible_tar_gzip vendor

  maindir=$PACKAGE-build
  mkdir -p $maindir
  mv vendor.tar.gz "${maindir}/"

  cd "${maindir}"

  for i in $sys; do
    os=$(echo $i | cut -f1 -d-)
    arch=$(echo $i | cut -f2 -d-)
    arm=

    if [[ $arch == "armv6" ]]; then
      arch=arm
      arm=6
    elif [[ $arch == "armv7" ]]; then
      arch=arm
      arm=7
    fi

    dir="${PACKAGE}-${i}"
    mkdir "${dir}"
    pushd "${dir}"

    green " - Building: ${os} ${arch} ${arm}"
    env CGO_ENABLED=0 GOOS=$os GOARCH=$arch GOARM=$arm go build -v -o ${PACKAGE} ${ldflags} ${package}
    popd

    # Add the hashes for the individual binaries as well for easy verification
    # of a single installed binary.
    sha256sum "${dir}/"* >> "manifest.txt" 

    if [[ $os == "windows" ]]; then
      reproducible_zip "${dir}"
    else
      reproducible_tar_gzip "${dir}"
    fi
  done

  # Add the hash of the packages too, then sort by the second column (name).
  sha256sum c13n-* vendor* >> "manifest.txt"
  LC_ALL=C sort -k2 -o "manifest.txt" "manifest.txt"
  cat "manifest.txt"
}

# usage prints the usage of the whole script.
function usage() {
  red "Usage: "
  red "release.sh build-release <version-tag> <build-system(s)> <build-tags> <ldflags>"
}

# Whatever sub command is passed in, we need at least 2 arguments.
if [ "$#" -lt 2 ]; then
  usage
  exit 1
fi


# Extract the sub command and remove it from the list of parameters by shifting
# them to the left.
SUBCOMMAND=$1
shift

# Call the function corresponding to the specified sub command or print the
# usage if the sub command was not found.
case $SUBCOMMAND in
build-release)
  green "Building release"
  build_release "$@"
  ;;
*)
  usage
  exit 1
  ;;
esac