## Running

### Running your own devnet

`--debugging --network="devnet" --new-devnet --tcp-server-port="5231" --set-genesis="file"  --staking`

#### Activating devnet faucet

`--hcaptcha-site-key="10000000-ffff-ffff-ffff-000000000001" --hcaptcha-secret="0x0000000000000000000000000000000000000000" --faucet-testnet-enabled="true"`

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
