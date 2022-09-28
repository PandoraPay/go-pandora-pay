## Running

### Running your own devnet

`--debugging --network="devnet" --new-devnet --tcp-server-port="5231" --set-genesis="file"  --forging`

#### Activating devnet faucet

hcaptcha site key must be set in /static/challenge/challenge.html
`--hcaptcha-secret="0x0000000000000000000000000000000000000000" --faucet-testnet-enabled="true"`

you can also create an account on hcaptcha

### Running the node as a Tor Hidden Server
1. Install Tor
2. Configure Tor
    - Ubuntu:
        - open `sudo nano /etc/tor/torrc`
        - add & save
            ``` 
            HiddenServiceDir /var/lib/tor/pandora_pay_hidden_service/
            HiddenServicePort 80 127.0.0.1:8080
            ```      
        - restart tor `sudo service tor restart`
        - copy your onion address `sudo nano /var/lib/tor/pandora_pay_hidden_service/`
        - use the parameter `--tor-onion="YOUR_ONION_ADDRESS_FROM_ABOVE"`

#### Running testnet script

`--run-testnet-script` will enable the testnet script which will create dummy transactions.

# DISCLAIMER:
This source code is released for research purposes only, with the intent of researching and studying a decentralized p2p network protocol.

PANDORAPAY IS AN OPEN SOURCE COMMUNITY DRIVEN RESEARCH PROJECT. THIS IS RESEARCH CODE PROVIDED TO YOU "AS IS" WITH NO WARRANTIES OF CORRECTNESS. IN NO EVENT SHALL THE CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES. USE AT YOUR OWN RISK.

You may not use this source code for any illegal or unethical purpose; including activities which would give rise to criminal or civil liability.

Under no event shall the Licensor be responsible for the activities, or any misdeeds, conducted by the Licensee.