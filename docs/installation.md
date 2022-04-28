## Installing

Go **1.18** is required as go-pandora-pay uses generics.


1. Install golang https://golang.org/doc/install
   1. For Linux
      1. `$ wget https://go.dev/dl/go1.18.1.linux-amd64.tar.gz`
      2. `$ sudo rm -rf /usr/local/go && tar -C /usr/local -xzf go1.18.1.linux-amd64.tar.gz`
      3. `$ sudo nano ~/.bashrc`
      4. Add at the bottom of the file
         ```
         export GOPATH=/usr/local/go
         export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
         ```
      5. `$ sudo source ~/.bashrc` 
2. Installing missing packages `go get -t .`
3. Run the node

### Checking and Installing a specific go version
1. `go env GOROOT`
2. download from https://golang.org/doc/manage-install

### Installing TLS/SSL Certificates

To install TLS certificates, you need to place the certificates in the application root folder with the following names
`certificate.crt`
`certificate.key`

You also need to specify the **domain address** by using the argument `--tcp-server-address="domain.net"`

To get authority certificates, you can use [cerbot](https://certbot.eff.org) (it's easy!) / or [Let's Encrypt](https://letsencrypt.org/) 

# DISCLAIMER:
This source code is released for research purposes only, with the intent of researching and studying a decentralized p2p network protocol.

PANDORAPAY IS AN OPEN SOURCE COMMUNITY DRIVEN RESEARCH PROJECT. THIS IS RESEARCH CODE PROVIDED TO YOU "AS IS" WITH NO WARRANTIES OF CORRECTNESS. IN NO EVENT SHALL THE CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES. USE AT YOUR OWN RISK.

You may not use this source code for any illegal or unethical purpose; including activities which would give rise to criminal or civil liability.

Under no event shall the Licensor be responsible for the activities, or any misdeeds, conducted by the Licensee.