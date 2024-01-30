package libhosty

// linuxHostsTemplate defines default linux hosts file
const linuxHostsTemplate = `# Do not remove the following line, or various programs
# that require network functionality will fail.
127.0.0.1       localhost.localdomain localhost
::1             localhost6.localdomain6 localhost6
`

// windowsHostsTemplate defines default windows hosts file
const windowsHostsTemplate = `# Copyright (c) 1993-2006 Microsoft Corp.
#
# This is a sample HOSTS file used by Microsoft TCP/IP for Windows.
#
# This file contains the mappings of IP addresses to host names. Each
# entry should be kept on an individual line. The IP address should
# be placed in the first column followed by the corresponding host name.
# The IP address and the host name should be separated by at least one
# space.
#
# Additionally, comments (such as these) may be inserted on individual
# lines or following the machine name denoted by a '#' symbol.
#
# For example:
#
#      102.54.94.97     rhino.acme.com          # source server
#       38.25.63.10     x.acme.com              # x client host
# localhost name resolution is handle within DNS itself.
#       127.0.0.1       localhost
#       ::1             localhost
`

// darwinHostsTemplate defines default darwin hosts file
const darwinHostsTemplate = `##
# Host Database
#
# localhost is used to configure the loopback interface
# when the system is booting. Do not change this entry
##
127.0.0.1           localhost
255.255.255.255     broadcasthost::1 localhost

::1                 localhost
fe80::1%lo0         localhost
`

// dockerDesktopTemplate defines docker desktop hosts entry
const dockerDesktopTemplate = `# Added by Docker Desktop
# To allow the same kube context to work on the host and the container:
127.0.0.1 kubernetes.docker.internal
# End of section
`
