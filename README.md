# lorona
A monitoring solution for servers and docker images. An exporter for prometheus.

# Install on Ubuntu
- Login to Ubuntu
- Install golang with 'sudo apt install golang-go'
- Run git clone git@github.com:markessien/lorona.git
- Enter the lorona directory with 'cd lorona'
- Type go get ./... to install all dependencies
- Type go build
- You now have an executable file called lorona
- Copy settings.sample.yaml to settings.yaml in the same directory
- Edit settings.yaml to point to monitor everything you are interested in
- Start lorona using ./lorona.
- If you need prometheus metrics, it listens on 2112 by default
- Open <ipaddress>:2112/metrics to get the prometheus metrics
- Create file /etc/systemd/system/lorona.service
- Paste this in:

```
[Unit]
Description=Lorona monitoring
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User=<your username>
ExecStart=/<path to lorona>/lorona -settings=/<path to lorona>/settings.yaml

[Install]
WantedBy=multi-user.target
```

- Start it with systemctl start lorona
- Test if it is running fine with ```systemctl status lorona```
- If all is fine, enable it with ```systemctl enable lorona```


# Notes
- There is a sample grafana dashboard in the repo

