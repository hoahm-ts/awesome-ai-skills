// Package event contains Kafka event publishers, consumers, and domain event types.
//
// Publishers send domain events to Kafka topics. Consumers subscribe to topics
// and invoke domain services. Both implement port interfaces defined in the relevant
// domain packages.
package event
