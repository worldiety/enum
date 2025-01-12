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

func ExampleDeclare() {
	type Pet interface {
		Eat()
		Sleep()
	}

	type Dog struct {
		TaxNumber int
		Pet
	}

	type Cat struct {
		Name string
		Pet
	}

	// usually declare it at package level
	// it is only for illustration here
	var PetEnum = enum.Declare[Pet, func(func(Dog), func(Cat), func(any))]()

	var myPet Pet
	myPet = Cat{Name: "Simba"}

	// Output: clean litterbox: Simba
	PetEnum.Switch(myPet)(
		func(dog Dog) {
			fmt.Printf("pay tax: %v\n", dog.TaxNumber)
		},
		func(cat Cat) {
			fmt.Printf("clean litterbox: %v\n", cat.Name)
		},
		func(a any) {
			if a != nil {
				fmt.Printf("remove vermin: %v\n", a)
			}
		},
	)
}

func ExampleNoZero() {
	type Pet interface {
		Eat()
		Sleep()
	}

	type Dog struct {
		TaxNumber int
		Pet
	}

	type Cat struct {
		Name string
		Pet
	}

	// usually declare it at package level
	// it is only for illustration here
	var PetEnum = enum.Declare[Pet, func(func(Dog), func(Cat))](
		enum.NoZero(),
	)

	var myPet Pet
	myPet = Dog{TaxNumber: 42}

	// Output: pay tax: 42
	PetEnum.Switch(myPet)(
		func(dog Dog) {
			fmt.Printf("pay tax: %v\n", dog.TaxNumber)
		},
		func(cat Cat) {
			fmt.Printf("clean litterbox: %v\n", cat.Name)
		},
	)
}

type petImpl struct {
}

func (p petImpl) Eat()   {}
func (p petImpl) Sleep() {}

func ExampleVariant() {
	type Pet interface {
		Eat()
		Sleep()
	}

	type Dog struct {
		TaxNumber int
		Pet
	}

	type Cat struct {
		Name string
		Pet
	}

	// usually declare it at package level
	// it is only for illustration here
	var _ = enum.Variant[Pet, Dog]()
	var _ = enum.Variant[Pet, Cat]()

	decl, ok := enum.DeclarationFor[Pet]()
	if !ok {
		panic("unreachable in this example")
	}

	// Output:
	// pet type: Dog
	// pet type: Cat
	for variant := range decl.Variants() {
		fmt.Printf("pet type: %s\n", variant.Name())
	}
}

func ExampleExternally() {
	type Pet interface {
		Eat()
		Sleep()
	}

	type Dog struct {
		TaxNumber int
		Pet
	}

	type Cat struct {
		Name string
		Pet
	}

	// usually declare it at package level
	// it is only for illustration here
	var _ = enum.Variant[Pet, Dog](
		// this is the default and can be omitted
		enum.Externally(),
	)
	var _ = enum.Variant[Pet, Cat]()

	var myPet Pet
	myPet = Cat{Name: "Simba"}
	// Note: to not lose the interface information, you MUST provide the pointer
	// to the interface variable, otherwise the type itself is marshalled.
	buf, err := json.Marshal(&myPet)
	if err != nil {
		panic(fmt.Errorf("unreachable in this example: %w", err))
	}
	// Output: {"Cat":{"Name":"Simba","Pet":null}}
	fmt.Println(string(buf))

	var pet2 Pet
	if err := json.Unmarshal(buf, &pet2); err != nil {
		panic(fmt.Errorf("unreachable in this example: %w", err))
	}

	if pet2 != myPet {
		panic(fmt.Errorf("unreachable in this example: %w", err))
	}
}

func ExampleAdjacently() {
	type Pet interface {
		Eat()
		Sleep()
	}

	type Dog struct {
		TaxNumber int
		Pet
	}

	type Cat struct {
		Name string
		Pet
	}

	// usually declare it at package level
	// it is only for illustration here
	var _ = enum.Variant[Pet, Dog](
		enum.Adjacently("kind", "obj"),
	)
	var _ = enum.Variant[Pet, Cat]()

	var myPet Pet
	myPet = Cat{Name: "Simba"}
	// Note: to not lose the interface information, you MUST provide the pointer
	// to the interface variable, otherwise the type itself is marshalled.
	buf, err := json.Marshal(&myPet)
	if err != nil {
		panic(fmt.Errorf("unreachable in this example: %w", err))
	}
	// Output: {"kind":"Cat","obj":{"Name":"Simba","Pet":null}}
	fmt.Println(string(buf))

	var pet2 Pet
	if err := json.Unmarshal(buf, &pet2); err != nil {
		panic(fmt.Errorf("unreachable in this example: %w", err))
	}

	if pet2 != myPet {
		panic(fmt.Errorf("unreachable in this example: %w", err))
	}
}
