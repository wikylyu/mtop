# MTop

`Enjoy the vision at mountain top.`

MTop is a simple network proxy. It's based on TLS with a very simple and understandable protocol.

## What it is and what it is not

MTop follows the principle -- `do one thing and do it well`.

MTop is just a network proxy for data forwarding. It provides below features:

* User auth with username/password
* Transport protocol is configurable. By default, it's TLS1.3, and [QUIC](https://en.wikipedia.org/wiki/QUIC) is supported. More transport protocol may be supported in the future, but they must be a security protocol.
* Custom CA. Allowed to use self signed certificate.
* MySQL and PostgreSQL integration for user management, which means you can just insert/update database to manage large scale of users programmatically, without modifying config file.

MTop doesn't and also will not provide below features:

* GeoIP detection and forwarding based on IP or domain.
* TCP Multiplexing. It should be implemented by application protocol, like HTTP/2.
* Complex encryption. MTop's safety depends on transport protocol.
* UDP forwarding.

## MTop Protocol

The initial handshake consists of the following:

1. Client connects and sends a authentication message, which includes client's username & password and the target domain/ip.
2. Server verifys client's username & password.
   1. If it's valid, try to connect to target domain/ip. When connection established, responses client with a success message. When connection establishment fails, responses client with a failure message.
   2. If it's invalid, responses client with a forbidden message.
3. Several messages may now pass between the client and the server.

### Client authentication message

|            | VER | METHOD | USERNAME | PASSWORD | ADDRESS  |
| ---------- | --- | ------ | -------- | -------- | -------- |
| Byte count | 1   | 1      | VARIABLE | VARIABLE | VARIABLE |

* **VER**
  MTop protocol version(0x1).

* **METHOD**
  Request method, only one method is supported now.
  * CONNECT(0x1)
  
* **USERNAME** & **PASSWORD**
  
    |            | LEN | TEXT |
    | ---------- | --- | ---- |
    | Byte count | 1   | LEN  |

    1 byte of name length followed by 1–255 bytes for the text.

* **ADDRESS**

    |            | TYPE | ADDR     |
    | ---------- | ---- | -------- |
    | Byte count | 1    | variable |

    * *TYPE*
        type of the address. One of:

            0x01: IPv4 address
            0x03: Domain name
            0x04: IPv6 address

    * *ADDR*
        the address data that follows. Depending on type:

            4 bytes for IPv4 address
            1 byte of name length followed by 1–255 bytes for the domain name
            16 bytes for IPv6 address

### Server response message

|            | VER | STATUS |
| ---------- | --- | ------ |
| Byte count | 1   | 1      |

* **VER**
  
  MTop protocol version(0x1).

* **STATUS**
  
  Response status code.

        0x0 Succcess
        0x1 Version not supported.
        0x2 Auth failure
        0x3 Connection failure.(Connecting to remote failure)
   
  
  In practise, the server will not response any message except *Success*, they just close the connection.


## Dataflow

```
+-----------------+                    +--------------------+                                 +----------------+
|  Application    |    HTTP/SOCKS5     |  Climber           |    MTOP Tunnel(over TLS/QUIC)   |  MTop Server   |
|  (Browser, etc) |  <-------------->  |  (Running locally) |  <----------------------------> |  (Remote)      |
+-----------------+                    +--------------------+                                 +----------------+
```

## Generate self-signed certificate

Replace example.com to your domain name or use IP:127.0.0.1 for testing.
    
```openssl genrsa -out ca.key 2048```

```openssl req -new -x509 -days 365 -key ca.key -subj "/C=CN/ST=GD/L=SZ/O=Acme, Inc./CN=Acme Root CA" -out ca.crt```

```openssl req -newkey rsa:2048 -nodes -keyout server.key -subj "/C=CN/ST=GD/L=SZ/O=Acme, Inc./CN=*.example.com" -out server.csr```

```openssl x509 -req -extfile <(printf "subjectAltName=DNS:example.com,DNS:www.example.com") -days 365 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt```

## Make & Install

### Install MTop
    
* install mtop to /usr/local/
  
  ```make install-mtop```
   
* install mtop to /custom/folder/

  ```PREFIX=/custom/folder/ make install-mtop```

* install mtop systemd script

  ```make install-mtop-systemd```

### Install Climber

*climber is the client for mtop service.*

* install climber to /usr/local/
  
  ```make install-climber```
   
* install climber to /custom/folder/

  ```PREFIX=/custom/folder/ make install-climber```

* install climber systemd script

  ```make install-climber-systemd```

After installing, you'll have to modify their config files to make them work.
