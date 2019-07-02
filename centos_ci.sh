#!/bin/bash

# Output command before executing
set -x

# Exit on error
set -e

########################################################
# Exit with message on failure of last executed command
# Arguments:
#   $1 - Exit code of last executed command
#   $2 - Error message
########################################################
function exit_on_failure() {
  if [[ "$1" != 0 ]]; then
    echo "$2"
    exit 1
  fi
}

# Source environment variables of the jenkins slave
# that might interest this worker.
function load_jenkins_vars() {
  if [ -e "jenkins-env" ]; then
    cat jenkins-env \
      | grep -E "(JENKINS_URL|GIT_BRANCH|GIT_COMMIT|BUILD_NUMBER|ghprbSourceBranch|ghprbActualCommit|BUILD_URL|ghprbPullId|CICO_API_KEY|GITHUB_TOKEN|JOB_NAME|RELEASE_VERSION|RH_REGISTRY_PASSWORD|CRC_BUNDLE_PASSWORD)=" \
      | sed 's/^/export /g' \
      > ~/.jenkins-env
    source ~/.jenkins-env
  fi

  echo 'CICO: Jenkins ENVs loaded'
}

function install_required_packages() {
  # Install EPEL repo
  yum -y install epel-release
  # Get all the deps in
  yum -y install make \
                 git \
                 curl \
                 kvm \
                 qemu-kvm \
                 libvirt \
                 python-requests \
                 libvirt-devel \
                 jq \
                 gcc \
		 golang \
		 libcurl-devel \
		 glib2-devel \
		 openssl-devel \
		 asciidoc \
		 unzip

  echo 'CICO: Required packages installed'
}

function setup_golang() {
  # Show which version of golang in the offical repo.
  go version
  # Setup GOPATH
  mkdir $HOME/gopath $HOME/gopath/src $HOME/gopath/bin $HOME/gopath/pkg
  export GOPATH=$HOME/gopath
  export PATH=$GOROOT/bin:$GOPATH/bin:$PATH
}

function perform_artifacts_upload() {
  set +x

  # For PR build, GIT_BRANCH is set to branch name other than origin/master
  if [[ "$GIT_BRANCH" = "origin/master" ]]; then
    # http://stackoverflow.com/a/22908437/1120530; Using --relative as --rsync-path not working
    mkdir -p crc/master/$BUILD_NUMBER/
    cp out/* crc/master/$BUILD_NUMBER/
    RSYNC_PASSWORD=$1 rsync -a --relative crc/master/$BUILD_NUMBER/ minishift@artifacts.ci.centos.org::minishift/crc/
    echo "Find Artifacts here http://artifacts.ci.centos.org/${REPO_OWNER}/${REPO_NAME}/master/$BUILD_NUMBER ."
  else
    # http://stackoverflow.com/a/22908437/1120530; Using --relative as --rsync-path not working
    mkdir -p pr/$ghprbPullId/
    cp -R out/* pr/$ghprbPullId/
    RSYNC_PASSWORD=$1 rsync -a --relative pr/$ghprbPullId/ minishift@artifacts.ci.centos.org::minishift/crc/
    echo "Find Artifacts here http://artifacts.ci.centos.org/minishift/crc/pr/$ghprbPullId ."
  fi
}

function get_bundle() {
  mkdir $HOME/Downloads
  curl -L 'https://github.com/code-ready/crc-ci-jobs/releases/download/4.0.1/crc_bundle.zip' -o $HOME/Downloads/bundle.zip
  unzip -P $CRC_BUNDLE_PASSWORD $HOME/Downloads/bundle.zip -d $HOME/Downloads/
}

function upload_logs() {
  set +x

  # http://stackoverflow.com/a/22908437/1120530; Using --relative as --rsync-path not working
  mkdir -p pr/$ghprbPullId/
  cp -R test/integration/out/test-results/* pr/$ghprbPullId/
  cp $HOME/.crc/crc.log pr/$ghprbPullId/crc_$(date '+%Y_%m_%d_%H_%M_%S').log
  RSYNC_PASSWORD=$1 rsync -a --relative pr/$ghprbPullId/ minishift@artifacts.ci.centos.org::minishift/crc/
  echo "Find Logs here: http://artifacts.ci.centos.org/minishift/crc/pr/$ghprbPullId ."
}

function run_tests() {
  set +e
  make integration BUNDLE_LOCATION="/root/Downloads/crc_libvirt_4.1.3.tar.xz"
  if [[ $? -ne 0 ]]; then
    upload_logs $1
    exit 1
  fi
}

# Execution starts here
load_jenkins_vars
install_required_packages
setup_golang
export TERM=xterm-256color
get_bundle

# setup to run integration tests
make
make fmtcheck
make cross
crc setup

# Retrieve password for rsync and run integration tests
CICO_PASS=$(echo $CICO_API_KEY | cut -d'-' -f1-2)
run_tests $CICO_PASS
perform_artifacts_upload $CICO_PASS

