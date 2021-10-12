#!/bin/bash

# bundle location
BUNDLE_VERSION=4.9.0-rc.7
BUNDLE=crc_libvirt_$BUNDLE_VERSION.crcbundle
GO_VERSION=1.15.13

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
  set +x
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
  yum -q -y install epel-release
  # Get all the deps in
  yum -q -y install make \
    git \
    curl \
    qemu-kvm \
    libvirt \
    libvirt-devel \
    jq \
    gcc \
    unzip \
    podman

  # Install the required version of golang
  curl -L -O https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz
  tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
  
  echo 'CICO: Required packages installed'
}

# Create a user which has NOPASSWD sudoer role
function prepare_ci_user() {

  groupadd -g 1000 -r crc_ci && useradd -g crc_ci -u 1000 crc_ci
  chmod +w /etc/sudoers && echo "crc_ci ALL=(ALL)    NOPASSWD: ALL" >> /etc/sudoers && chmod -w /etc/sudoers

  # Copy centos_ci.sh to newly created user home dir
  cp centos_ci.sh /home/crc_ci/
  mkdir /home/crc_ci/payload
  # Copy crc repo content into crc_ci user payload directory for later use
  cp -R /root/payload /home/crc_ci/
  chown -R crc_ci:crc_ci /home/crc_ci/payload
  # Copy the jenkins-env into crc_ci home dir
  cp ~/.jenkins-env /home/crc_ci/jenkins-env

  loginctl enable-linger crc_ci
}

function setup_golang() {
  export PATH=$PATH:/usr/local/go/bin

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
    cp -R out/* crc/master/$BUILD_NUMBER/
    RSYNC_PASSWORD=$1 rsync -a --relative crc/master/$BUILD_NUMBER/ minishift@artifacts.ci.centos.org::minishift/crc/
    echo "Find Artifacts here http://artifacts.ci.centos.org/minishift/crc/master/$BUILD_NUMBER ."
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
  curl -L "https://storage.googleapis.com/crc-bundle-github-ci/crc_libvirt_$BUNDLE_VERSION.zip" -o $HOME/Downloads/bundle.zip
  unzip -P $CRC_BUNDLE_PASSWORD $HOME/Downloads/bundle.zip -d $HOME/Downloads/
}

function upload_logs() {
  set +x

  # http://stackoverflow.com/a/22908437/1120530; Using --relative as --rsync-path not working
  mkdir -p pr/$ghprbPullId/
  cp -R test/e2e/out/test-results/* pr/$ghprbPullId/
  # Change the file permission to 0644 so after rsync it can be readable by other user.
  chmod 0644 $HOME/.crc/crc.log
  cp $HOME/.crc/crc.log pr/$ghprbPullId/crc_$(date '+%Y_%m_%d_%H_%M_%S').log
  RSYNC_PASSWORD=$1 rsync -a --relative pr/$ghprbPullId/ minishift@artifacts.ci.centos.org::minishift/crc/
  echo "Find Logs here: http://artifacts.ci.centos.org/minishift/crc/pr/$ghprbPullId ."
}

function run_tests() {
  set +e
  # In Jenkins slave we have pull secret file in the $HOME/payload/crc_pull_secret
  # this is copied over using https://github.com/minishift/minishift-ci-jobs/blob/master/minishift-ci-index.yaml#L99
  export PULL_SECRET_FILE=--pull-secret-file=$HOME/payload/crc_pull_secret
  export BUNDLE_LOCATION=--bundle-location=$HOME/Downloads/$BUNDLE 
  make e2e 
  if [[ $? -ne 0 ]]; then
    upload_logs $1
    exit 1
  fi
}

# Execution starts here

if [[ "$UID" = 0 ]]; then
	load_jenkins_vars
	install_required_packages
	prepare_ci_user
	runuser -l crc_ci -c "/bin/bash centos_ci.sh"
else
	source ~/jenkins-env # Source environment variables for minishift_ci user
	export XDG_RUNTIME_DIR="/run/user/$UID"
	export DBUS_SESSION_BUS_ADDRESS="unix:path=${XDG_RUNTIME_DIR}/bus"
	export TERM=xterm-256color
	get_bundle
	setup_golang

	# setup to run e2e tests
	cd payload
	make
	make check
	
	# Retrieve password for rsync and run e2e tests
	CICO_PASS=$(echo $CICO_API_KEY | cut -d'-' -f1-2)
	run_tests $CICO_PASS
	perform_artifacts_upload $CICO_PASS
fi

