# Add Docker's official GPG key:
sudo apt-get update
sudo apt-get install -y ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

# Add the repository to Apt sources:
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update

sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

sudo apt install -y wget
wget -O /tmp/homecloud.tar.gz https://github.com/An-Owlbear/homecloud/releases/download/v0.1.2/homecloud-v0.1.2-linux-arm64.tar.gz
mkdir -p /opt/homecloud
tar -xf /tmp/homecloud.tar.gz -C /opt/homecloud

cp homecloud.service /etc/systemd/system
systemctl daemon-reload
systemctl enable homecloud
systemctl start homecloud