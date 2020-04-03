# Flashforge Finder proxy

This software proxies tcp and http connections.

```bash
go run . -l 0.0.0.0:8080 -t 192.168.1.12:8080

...
TCP Listening: 0.0.0.0:8899
TCP Proxying: 192.168.4.13:8899
HTTP Listening: http://0.0.0.0:9090)
HTTP Proxying: http://192.168.4.13/

http is statically configured to proxy:9090 -> target:80
```

### Reason 1
I had problems connection my FlashForge Finder to my wifi router.   
Connecting the printer to a Raspberry Pi Access Point did work well.  

So I wrote this proxy to enable to connecting my pc over wifi over pi wifi with the printer.

Now I can send a print job directly from my pc without using usb stick.

### Reason 2
It's fun