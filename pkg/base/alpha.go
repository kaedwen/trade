package base

import (
	"fmt"

	alpha "github.com/kaedwen/go-alpha-vantage"
	"github.com/rs/zerolog/log"
)

var alphaClient *alpha.Client
var symbolMap = map[string]string{}

func setupAlpha() {

	connection := alpha.NewConnection()
	alphaClient = alpha.NewClientConnection(Config.AlphaVantage.ApiKey, connection)

}

func lookupSymbol(name string) (string, error) {
	lookup := func(name string) (string, error) {
		if value, ok := symbolMap[name]; ok {
			return value, nil
		} else {
			if symbols, err := alphaClient.SymbolSearch(name); err == nil {
				if len(symbols.BestMatches) > 0 {
					symbol := (*symbols.BestMatches[0]).Symbol
					symbolMap[name] = symbol
					return symbol, nil
				}
			}
		}
		return "", fmt.Errorf("failed to lookup symbol for name %s", name)
	}

	symbol, err := lookup(name)
	if err == nil {
		log.Info().Msgf("Symbol for %s is %s", name, symbol)
	}
	return symbol, err
}

func GetQuote(target string) (Quote, error) {

	symbol, err := lookupSymbol(target)
	if err != nil {
		return Quote{}, err
	}

	quote, err := alphaClient.Quote(symbol)
	if err != nil {
		return Quote{}, err
	}

	log.Info().Msgf("Symbol %s with price %f", symbol, quote.Quote.Price)

	return Quote{
		Symbol: symbol,
		Value:  quote.Quote.Price,
	}, nil

}
