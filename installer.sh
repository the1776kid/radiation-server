#!/bin/bash
program_name="radiation-server"
service_file="/etc/systemd/system/${program_name}.service"
storage_directory="/usr/share/radiation-server/"
if (( $EUID == 0 )); then
  echo "Please run as user. Thank you!!"
  exit
fi
user_name=$(whoami)
function build() {
  CGO_ENABLED=0 go build -v -o ${program_name} -ldflags="-s -w" main.go
}
function installService() {
  echo "[Unit]
Description=${program_name}
After=network.target

[Service]
Type=simple
User=${user_name}
ExecStart=/usr/bin/${program_name}
Restart=always

[Install]
WantedBy=multi-user.target"  > ${program_name}.service
  sudo mv ${program_name}.service /etc/systemd/system/
  sudo systemctl enable ${program_name}.service
  sudo systemctl start ${program_name}.service
}
function installBinary() {
  sudo mv ${program_name} /usr/bin/
}
function install() {
  if [ ! -d "${storage_directory}" ]; then
    echo "${storage_directory} does not exist."
    sudo mkdir /usr/share/radiation-server/
    sudo chmod 777 /usr/share/radiation-server/
  else
    echo "${storage_directory} exists."
  fi
  build
  installBinary
  installService
}
function update() {
  if [[ -f "${service_file}" ]]; then
    sudo systemctl stop ${program_name}.service
    build
    installBinary
    sudo systemctl start ${program_name}.service
  else
    echo "${program_name} not installed
    bash installer.sh install"
  fi
}
case "$1" in
  "install")
    install;;
  "update")
    update;;
  "build")
    build;;
  *)
    printf "%s\n" "bash installer.sh <arg>
 install       build , install binary, install and enable service
 update        rebuild , install binary, restart service
 build         build binary
    ";;
esac
