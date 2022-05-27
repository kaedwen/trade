# Comdirect Trade App

This app connects to the comdirect api to query your contract with the give credentials. Pulling the current account balances, depot values and pushing them to the configured influx ORG/Bucket.

## Config
Configuration is looked up in the following manner
* .config.yaml from `cwd` up to /
* `$HOME`/.config/trade/config.yaml
* /etc/trade/config.yaml

Please fill in your crendetials in any location (first wins, no merge)
