# Proto-type Heterodoxy

(NOTE: all heterodxy relative to standard patterns and best practices in related projects)

This document also flirts with unabashed blasphemy where appropriate.

TLDR: Document and explain deviations, divergences, generalizations, tweaks, and disagreements from best practices and convention in related projects such as

* Protobuf and gRPC

* [Google AIP](https://google.aip.dev/general)

* [Google Service Control](https://docs.cloud.google.com/service-infrastructure/docs/service-control/getting-started)

* [Google API/resource model and service platform impl](https://github.com/googleapis/googleapis/tree/master/google/api)

* [GCP's managed/prescribed gRPC platform Cloud Endpoints](https://docs.cloud.google.com/endpoints/docs/grpc/about-grpc)

* [Protobuf's reflection model](https://github.com/protocolbuffers/protobuf/blob/main/src/google/protobuf/descriptor.proto) and related packages/libraries 

* Proto/grpc tooling and ecosystem projects such as [grpc-web](https://github.com/grpc/grpc-web)

* Golang and its ecosystem

* Buf's understanding of schema's, "registries", etc eg [buf](https://github.com/bufbuild/buf) or [Buf Schema Registry](https://buf.build/docs/bsr/)

* General type, distributed systems, security, agent orchestration practices.

# AIP

Google's massive corpus of "API Improvement Proposals" is a clearly well-intentioned attempt at solving what should be a straightforward human-coordination problem with a death-by-a-thousand-cuts rulebook of no-no's, requirements, and strongly suggested proscriptive behavior. At a small scale, having an API design rule book makes sense, but AIP is about a hundred pages of rules like

* Collection identifiers must be in camelCase.

* Field masks should not be specified in the request.

* Resources with a revision history may have child resources. If they do, there are two potential variants

I hate this. Excuse me, this is an API design guide deeply intertwined with a type system: **GOOD INTERFACES MAKE INVALID STATES UNREPRESENTABLE**

How can a veritable bible of proscribed rules for APIs fail so completely at recognizing this? This is precisely the problem that type systems and good design solve. All this does is inflict OCPD and poor problem:solution coherence on someone who's actually trying to do real work, and now has to battle a linter, API READABILITY, launch approval, and follow every rule across some enormous design-by-committee binder of infrastructural metaphysics?

Anyway, there are many Problems with this system. But, some of the ideas and concepts could be useful IF they were properly implemented in a proper type system, that would make disallowed interfaces unrepresentable or invalid by default, provide Interfaces for establishing classes with shared structure independently of encapsulation, and use derived types (eg Generic, Traits) for implementing common, higher-order patterns or simple type enrichment.

I don't even want to credit some of those common structures to AIP, because most are so obvious or standardized even outside of Google/Proto that I was planning on doing them before I realized AIP had Special Rules for them.

I'll focus more on where we diverge in our understandings of common structures and interfaces. But, if you butcher the API design of an API design system, you probably need extra help even explaining what the purpose or benefits of this stuff is to begin with, so I'll help with that or play devil's advocate too where appropriate.

## Terminology

Concepts and terminology I like:

* API [planes](https://google.aip.dev/111). **Data plane** and **control plane**, not management plane.

These are fundamentally different kinds of interfaces for interacting with remote state: the control plane operates on persistent (usually to a database) identities and "resources", typically has lower availability/throughput/latency requirements, is a managed remote service in a different trust domain. The data plane has higher availability/throughput/performance requirements and is managed remotely but exists in a shared/client-owned trust domain.

* Resources. Resources are not "nouns", they are representations of persistent, remote state.

The control plane deals in resources: abstract, compact, stateful representations of things like "virtual machines" and "authorized users" that get stored in databases and are expected not to change their resources' state in the absence of further direct control plane operations on the resource. Conversely, the data plane deals with the **actual** resource, which may not even have a state that can be serialized into a self-consistent snapshot or fully read by a remote client.

There are many analogies you could draw to help better explain this. Let's go with one that is on-topic to type system design: as we all know, [the map is not the territory](https://en.wikipedia.org/wiki/Map%E2%80%93territory_relation), and although maps are still damn useful if you want to understand the territory, not only can they be wrong or misleading or out of date, **there is not even necessarily between a maps' representation of a territory, and the actual territory in question, or even reality itself**.

Then, resources are the remotely managed "maps" that describes the essential, fixed metadata handled by the control plane, that helps explain the much-harder-to-describe entities they actually represent. From day to day, maps don't meaningfully change. But the territory is always changing. Similarly, a database's disk size is a constant until someone or something deliberately provisions or adds additional disk space to it; but the *contents* of the disk, and the amount of data actively used, and the number of readers in this very moment, and how many database tables it has on it, are all things that could be changed through active interaction with the database itself through the dataplane.

Another way to look at it, is that control plane operations that directly mutate resources - creating, deleting, or updating them - typically involve some kind of cost implication. These may be significant, and continue outside the scope of any other control plane operations (data plane operations also often are modeled as involving costs, but less oftenly and typically per-call) at some kind of rate, eg $0.10/hr. Also, to have a resource that is interactable on the dataplane, you or someone else usually needs to have provisioned it first, so "all of my resources" roughly models "all the computational entities I am responsible for/all the costly computational entities I am paying for". In that way, resources are like the "deed" to a house, or a treaty, or a subscription, and dataplane costs are like either eating at an all-you-can-eat buffet, or paying a toll-road, or going to a restaurant and leaving with no food but an empty belly.

There is one crucial, subtle, often-overlooked detail about resources that even most experts tend to miss: resources merely encode point-in-time metadata about the actual referenced entity not just to end-user clients, *but also* the control plane, *and* the database. All of these can disagree with each other, and also be a wrong or (almost always) incomplete representation of the *actual* thing the resource references. In that way, a resource is like a contract or receipt: it describes what you're *paying for*, not what you *bought*.

* Resource-Oriented Design

Resources are not just "nouns" - see above. More importantly, we can just implement generics and not need "typical resource APIs". We've basically already done that in https://github.com/accretional/collector

### Resources are Great, Messages are Good, Request/Response types are overrated

Oftentimes, a gRPC service's methods will each be given their own custom request and response types for each method. Eg

```proto

message FooCommon {
  string name = 1;
}

message FooRequest {
  FooCommon f = 1;
}

message FooResponse {
  FooCommon f = 1;
}

service MyService {
  rpc Foo(FooRequest) returns(FooResponse);
}
```

In many ways this is just unnecessary boilerplate around what could be a much more useful type system. But it's also an important way of future-proofing against types that may not be different now, but represent fundamentally different/decoupled things that *may* be different in the future.
  
