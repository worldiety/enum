// Copyright 2024 Torben Schinke. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package enum_test

import (
	"fmt"
	"github.com/worldiety/enum"
	"github.com/worldiety/enum/json"
	"testing"
)

var OfferEnum = enum.Declare[Offer, func(func(AcceptedOffer), func(UncheckOffer), func(any))](
	//enum.Rename[WdyCoin]("wdy-coin"),
	//	enum.Adjacently("t2", "c2"),
	//enum.NoZero(),
	enum.Sealed(),
)

type Offer interface {
	offer()
}
type AcceptedOffer struct {
	Sum Currency
}

func (AcceptedOffer) offer() {}

type UncheckOffer struct {
	Sum Currency
}

func (UncheckOffer) offer() {}

var CurrencyEnum = enum.Declare[Currency, func(func(Dollar), func(EuroCent), func(FakeMoney), func(any))](
	enum.Adjacently("t3", "c3"),
)

type Currency interface {
	currency()
}
type Dollar int

func (Dollar) currency() {}

type EuroCent int64

func (EuroCent) currency() {}

type FakeMoney interface {
	fakeMoney()
	currency()
}

type WdyCoin struct{}

func (WdyCoin) fakeMoney() {}

type WizCoin [32]byte

func (WizCoin) fakeMoney() {}

type UnsealedOffer struct{}

func (UnsealedOffer) offer() {}

func TestDeclare2(t *testing.T) {

	var offer Offer
	//offer = UnsealedOffer{}
	offer = AcceptedOffer{Sum: EuroCent(2)}

	OfferEnum.Switch(offer)(
		func(offer AcceptedOffer) {
			fmt.Printf("acceppted offer: %v\n", offer)
			CurrencyEnum.Switch(offer.Sum)(func(dollar Dollar) {
				fmt.Printf("dollar: %v\n", dollar)
			}, func(cent EuroCent) {
				fmt.Printf("eurocent: %v\n", cent)
			}, func(money FakeMoney) {
				fmt.Printf("fake money: %v\n", money)
			}, func(a any) {

			})
		}, func(offer UncheckOffer) {
			fmt.Printf("unchecked offer: %v\n", offer)
		}, func(a any) {
			fmt.Printf("any offer: %v %T\n", offer, offer)
		})

	buf, err := json.Marshal(&offer)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(buf))

	var test Offer
	if err := json.Unmarshal(buf, &test); err != nil {
		t.Fatal(err)
	}

	if test != offer {
		t.Fatal("not equal")
	}

	fmt.Printf("%#v\n", test)
}
