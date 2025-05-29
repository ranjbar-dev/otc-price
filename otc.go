package main

import (
	"math"
	"math/rand/v2"
	"time"
)

type Otc struct {
	price                 float64
	generatedPrice        float64
	timestamp             int64
	historicalPrices      []float64
	historicalPricesLimit int
	upperBound            float64
	lowerBound            float64
	moveStrength          float64
	spikeProbability      float64
	spikeStrength         float64
	direction             float64
}

func (o *Otc) SetAndGeneratePrice(price float64) {

	// set price
	o.price = price

	o.randomizeConfigs()

	o.historicalPrices = append(o.historicalPrices, price)
	if len(o.historicalPrices) < o.historicalPricesLimit { // if we dont have enough historical prices set generated price to price and return

		o.generatedPrice = price
		return
	}

	// pop oldest price from historical prices
	o.historicalPrices = o.historicalPrices[1:]

	// define temp variable for storing generated price
	var temp float64

	// calculate volatility
	volatility := o.CalculateVolatility()

	if rand.Float64() < o.spikeProbability { // if spike happens we will go back to the real price

		if o.generatedPrice > o.price {

			temp = o.generatedPrice - (volatility * o.direction * o.spikeStrength)
		} else {

			temp = o.generatedPrice + (volatility * o.direction * o.spikeStrength)
		}

	} else { // generate normal price

		temp = o.generatedPrice + (volatility * o.moveStrength * o.direction)
	}

	// adjust to upper and lower bounds
	if temp > o.upperBound {

		temp = o.upperBound
	} else if temp < o.lowerBound {

		temp = o.lowerBound
	}

	// set generated price
	o.generatedPrice = temp

	// set timestamp
	o.timestamp = time.Now().Unix()
}

func (o *Otc) CalculateVolatility() float64 {

	n := float64(len(o.historicalPrices))
	var sum, mean, variance float64
	for _, price := range o.historicalPrices {

		sum += float64(price)
	}

	mean = sum / n
	for _, price := range o.historicalPrices {

		variance += math.Pow(float64(price)-mean, 2)
	}

	return math.Sqrt(variance / (n - 1))
}

func (o *Otc) randomizeConfigs() {

	// every 10 seconds
	if time.Now().Unix()%10 == 0 {

		// random number between 0.0 and 0.5
		random := rand.Float64() * 0.05

		// generate upper bound 1.00 to 1.10 real price
		o.upperBound = o.price * (random + 1.05)

		// generate lower bound 0.90 to 1.00 real price
		o.lowerBound = o.price * (random + 0.95)
	}

	// random number between 0.000 and 0.01
	o.spikeProbability = rand.Float64() * 0.01

	// random number between 1 and 2
	o.spikeStrength = rand.Float64()*1 + 1

	// random number between 0.0 and 1
	o.moveStrength = rand.Float64()

	// random direction
	if rand.Float64() < 0.5 {

		o.direction = 1
	} else {
		o.direction = -1
	}
}

func NewOtc(price float64, timestamp int64) *Otc {

	return &Otc{
		price:                 price,
		generatedPrice:        price,
		historicalPrices:      []float64{price},
		historicalPricesLimit: 60,
		upperBound:            price,
		lowerBound:            price,
		moveStrength:          1.00,
		spikeProbability:      0.001,
		spikeStrength:         2.00,
		timestamp:             timestamp,
		direction:             1,
	}
}
