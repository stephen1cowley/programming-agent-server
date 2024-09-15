### To Run

Works on linux only.

Clone with http
```
git clone https://github.com/stephen1cowley/go-agent
```


Grant executable permissions for the shellscript
```
chmod +x go-agent/shell_script/*
```

Then
```
sudo apt-get install npm
npx create-react-app my-react-app
cd my-react-app
npm init -y
```

Then we want to get the server up and running on public IPv4 address. So in `package.json` under scripts:start prepend `HOST=0.0.0.0` and allow external http traffic on port 3000 on your server.

Now get the server running with an automatic restart daemon using `systemd`.

```
sudo nano /etc/systemd/system/my-react-app.service
```

Replace contents with (replace `ubuntu` with actual system username)
```
[Unit]
Description=My React App
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/my-react-app
ExecStart=/usr/bin/npm start
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Then in `my-react-app`
```
sudo systemctl daemon-reload
sudo systemctl start my-react-app
sudo systemctl status my-react-app.service
```

Now the service is up and running. 

Next, Add a `.env` file with `OPEN_AI_KEY = "abcdefghijklmnop"`

We're ready for our chatbot to make changes to the code! To run the chat bot:

```
sudo apt-get install golang
cd ~/go-agent
go run main.go
```