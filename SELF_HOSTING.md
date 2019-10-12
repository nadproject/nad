# Hosting NAD On Your Machine

This guide documents the steps for installing the NAD server on your own machine.

## Installation

1. Install Postgres 10+.
2. Create a `nad` database by running `createdb nad`
3. Download the official NAD server release from the [release page](https://github.com/nadproject/nad/releases).
4. Extract the archive and move the `nad-server` executable to `/usr/local/bin`.

```bash
tar -xzf nad-server-$version-$os.tar.gz
mv ./nad-server /usr/local/bin
```

4. Run NAD

```bash
GO_ENV=PRODUCTION \
DBHost=localhost \
DBPort=5432 \
DBName=nad \
DBUser=$user \
DBPassword=$password \
  nad-server start
```

Replace $user and $password with the credentials of the Postgres user that owns the `nad` database.

By default, nad server will run on the port 3000.

## Configuration

By now, NAD is fully functional in your machine. The API, frontend app, and the background tasks are all in the single binary. Let's take a few more steps to configure NAD.

### Configure Nginx

To make it accessible from the Internet, you need to configure Nginx.

1. Install nginx.
2. Create a new file in `/etc/nginx/sites-enabled/nad` with the following contents:

```
server {
	server_name my-nad-server.com;

	location / {
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $remote_addr;
		proxy_set_header Host $host;
		proxy_pass http://127.0.0.1:3000;
	}
}
```
3. Replace `my-nad-server.com` with the URL for your server.
4. Reload the nginx configuration by running the following:

```
sudo service nginx reload
```

Now you can access the NAD frontend application on `/`, and the API on `/api`.

### Configure TLS by using LetsEncrypt

It is recommended to use HTTPS. Obtain a certificate using LetsEncrypt and configure TLS in Nginx.

In the future versions of the NAD Server, HTTPS will be required at all times.

### Run NAD As a Daemon

We can use `systemd` to run NAD in the background as a Daemon, and automatically start it on system reboot.

1. Create a new file at `/etc/systemd/system/nad.service` with the following content:

```
[Unit]
Description=Starts the nad server
Requires=network.target
After=network.target

[Service]
Type=simple
User=$user
Restart=always
RestartSec=3
WorkingDirectory=/home/$user
ExecStart=/usr/local/bin/nad-server start
Environment=GO_ENV=PRODUCTION
Environment=DBHost=localhost
Environment=DBPort=5432
Environment=DBName=nad
Environment=DBUser=$DBUser
Environment=DBPassword=$DBPassword
Environment=SmtpHost=
Environment=SmtpUsername=
Environment=SmtpPassword=

[Install]
WantedBy=multi-user.target
```

Replace `$user`, `$DBUser`, and `$DBPassword` with the actual values.

Optionally, if you would like to send email digests, populate `SmtpHost`,  `SmtpUsername`, and `SmtpPassword`.

2. Reload the change by running `sudo systemctl daemon-reload`.
3. Enable the Daemon  by running `sudo systemctl enable nad`.`
4. Start the Daemon by running `sudo systemctl start nad`

### Enable Pro version

After signing up with an account, enable the pro version to access all features.

Log into the `nad` Postgres database and execute the following query:

```sql
UPDATE users SET cloud = true FROM accounts WHERE accounts.user_id = users.id AND accounts.email = '$yourEmail';
```

Replace `$yourEmail` with the email you used to create the account.

### Configure clients

Let's configure NAD clients to connect to the self-hosted web API endpoint.

#### CLI

We need to modify the configuration file for the CLI. It should have been generated at `~/.nad/nadrc` upon running the CLI for the first time.

The following is an example configuration:

```yaml
editor: nvim
apiEndpoint: https://api.getnad.com
```

Simply change the value for `apiEndpoint` to a full URL to the self-hosted instance, followed by '/api', and save the configuration file.

e.g.

```yaml
editor: nvim
apiEndpoint: my-nad-server.com/api
```
