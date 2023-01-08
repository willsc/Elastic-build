#!/bin/bash

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    -u|--user)
      user="$2"
      shift
      shift
      ;;
    -p|--password)
      password="$2"
      shift
      shift
      ;;
    -s|--sftp-server)
      sftp_server="$2"
      shift
      shift
      ;;
    -x|--proxy-server)
      proxy_server="$2"
      shift
      shift
      ;;
    -P|--proxy-port)
      proxy_port="$2"
      shift
      shift
      ;;
    -r|--remote-dir)
      remote_dir="$2"
      shift
      shift
      ;;
    -l|--local-dir)
      local_dir="$2"
      shift
      shift
      ;;
    -t|--time-format)
      time_format="$2"
      shift
      shift
      ;;
    -m|--max-age)
      max_age="$2"
      shift
      shift
      ;;
    -b|--bandwidth)
      bandwidth="$2"
      shift
      shift
      ;;
    -f|--filename-pattern)
      filename_pattern="$2"
      shift
      shift
      ;;
    *)
      echo "Invalid argument: $key"
      exit 1
      ;;
  esac
done

# Set default values for variables if they are not set
user=${user:-anonymous}
password=${password:-}
sftp_server=${sftp_server:-localhost}
proxy_server=${proxy_server:-}
proxy_port=${proxy_port:-3128}
remote_dir=${remote_dir:-.}
local_dir=${local_dir:-.}
time_format=${time_format:-%Y-%m-%d %H:%M:%S}
max_age=${max_age:-0}
bandwidth=${bandwidth:-0}
filename_pattern=${filename_pattern:-*}

# Check if all required variables are set
if [[ -z "$sftp_server" ]]; then
  echo "Error: SFTP server not specified. Use the -s option to set the SFTP server."
  exit 1
fi

# Construct the sftp command
sftp_cmd="sftp -o \"BatchMode=no\""

# Add the ProxyCommand option if a proxy server is specified
if [[ -n "$proxy_server" ]]; then
  sftp_cmd+=" -o \"ProxyCommand nc -X connect -x $proxy_server:$proxy_port %h %p\""
fi

# Add the user and SFTP server to the sftp command
sftp_cmd+=" $user@$sftp_server"

# Connect to the SFTP server and download the latest files
echo "Connecting to $sftp_server..."
$sftp_cmd <<EOF
cd $remote_dir
mget -E $filename_pattern
exit

