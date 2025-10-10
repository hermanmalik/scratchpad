# shared scratchpad
This is really just so I can get links and stuff around.
It requires no auth and obviously has no guarantees as to the integrity or anonymity of the data you dump in. But it's fun!

## setup (nginx + systemd version)
- Install Go on the server
- Make /var/www/scratchpad, paste in `main.go` and `index.html`, then edit `main.go` with your website URL and port
- Build and execute the backend
    - `go mod init scratchpad` to initialize
    - `go get github.com/gorilla/websocket` for the dependency
    - `go build -o scratchpad` to build the binary
    - `chmod +x scratchpad` to be able to run it
- Make a systemd unit `/etc/systemd/system/scratchpad.service` and paste in the content of `scratchpad.service`, then enable and start it
- Check for SELinux issues
    - if `systemctl status scratchpad` shows it's not running, try adding `sudo semanage fcontext -a -t bin_t "/var/www/scratchpad/scratchpad"` and `sudo restorecon -v /var/www/scratchpad/scratchpad` to get SELinux to allow the service to run. 
    - I might also have had to run `sudo setsebool -P httpd_can_network_connect 1` to let the service talk to nginx (this error comes after setting up the proxy below)
- To proxy nginx traffic from the webserver to the scratchpad, put this block nested in the server block corresponding to your URL (make sure to edit "[PORT]")!
    ```
    location /scratchpad {
        proxy_pass http://localhost:[PORT]/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    location /scratchpad/ws {
        proxy_pass http://localhost:[PORT]/ws;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 86400;
    }
    ```
- Then run `sudo nginx -t` to test the config, and if there's nothing wrong `sudo systemctl reload nginx`
- After this you should be able to visit [your URL]/scratchpad to see it working
    - If you want the scratchpad on a different URL change both the nginx config and the javascript line setting `const wsUrl`
