# Comdirect Trade App

This app connects to the comdirect api to query your contract with the give credentials. Pulling the current account balances, depot values.

## Config
Configuration is looked up in the following manner
* .config.yaml next to bonary
* `$HOME`/.config/trade/config.yaml
* /etc/trade/config.yaml

Please fill in your crendetials in any location (first wins, no merge)

## Runtime
This project is based on systemd and provides `trade.service`

Please do not enable this service because you have to approve the TAN challenge every time the service ist started. So startup remains a manual process where you have to approve the TAN request in time.

