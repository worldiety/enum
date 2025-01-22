# enum

[![Go Reference](https://pkg.go.dev/badge/github.com/worldiety/enum.svg)](https://pkg.go.dev/github.com/worldiety/enum)

Package enum provides helpers to enable enumerable interfaces for Go.
Sometimes, exhaustive checks and enumeration are required to simplify code and let this fact be checked by the compiler.
In general, due to the way, the Go type system has been designed, this is not possible.
However, for non-general cases, this can be done under specific conditions with limitations.

For further discussions and backgrounds see [5468](https://github.com/golang/go/issues/54685) et al.

Our approach is novel in a way, that we don't use code generation or pre-created tuples or type parameters.
Instead, it combines a functional approach and reflection to delegate the checks either to the compiler or to the runtime at package initialization time. The actual matching logic is also implemented using reflection.

To allow seamless interoperability with JSON encoding and decoding, we forked the stdlib json package to directly support the enumerable interface types.

## sealed example

```go
//...

import(
    "github.com/worldiety/enum"
    "github.com/worldiety/enum/json"
)

// Note, that AcceptedOffer and UncheckOffer must be assignable to Offer, which is checked at 
// package initialization time. Offer must not be the empty interface.
var OfferEnum = enum.Declare[Offer, func(func(AcceptedOffer), func(UncheckOffer), func(any))](
	// options over options, inspired by and compatible with https://serde.rs/enum-representations.html
	enum.Rename[AcceptedOffer]("aof"),    // provide custom names
	enum.Adjacently("t", "c"),            // default is externally tagged, like serde
	//enum.NoZero(),                      // panic, if switch finds a zero-interface, you can omit the func(any) branch
	enum.Sealed(),                        // do not accept future Variant declaration (see second example below)
)

// ...

func someFn(){
    var offer Offer
	
	// now we can use the type enum switch func which has been implemented by reflection above.
	// if the enum is changed, this will fail at compile-time, and we can be sure to be
	// exhaustive with respect to our declaration. There is still nil and arbitrary other types,
	// but we can express that these types are essential for our domain and each case has been handled.
	OfferEnum.Switch(offer)(
        func(offer AcceptedOffer) {
            fmt.Printf("acceppted offer: %v\n", offer)
		}, 
		func(offer UncheckOffer) {
            fmt.Printf("unchecked offer: %v\n", offer)
        }, 
		func(a any) {
            fmt.Printf("any offer: %v %T\n", offer, offer)
        },
	)
	
	// ...
	// encode/decode as usual, but note the different import
    buf, err := json.Marshal(&offer)
    if err != nil {
        t.Fatal(err)
    }
    
    fmt.Println(string(buf))
}
```

## open type example

Sometimes, you just want to define a base interface in a supporting package, which others need to extend, which the supporting package needs to inspect.
We can model this situation as follows:

```go
//...

import(
    "github.com/worldiety/enum"
)

type Credentials interface {
    GetName() string
    Credentials() bool // open sum type which can be extended by anyone
    IsZero() bool
}

// declaring the variants associates the concrete types with the interface type
var (
    _ = enum.Variant[secret.Credentials, secret.Jira]()
    _ = enum.Variant[secret.Credentials, secret.BookStack]()
)

// you can even inspect the declared variants at runtime.
// note, that this cannot be exhaustive, but we find that to be reasonable enough
func someFn(){
    decl, ok := enum.DeclarationFor[secret.Credentials]()
    if !ok {
        panic("unreachable: secret.Credentials declaration not defined")
    }
    
    for rtype := range decl.Variants() {
		//...
    }
}
```
