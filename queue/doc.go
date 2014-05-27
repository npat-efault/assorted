// Package queue provides three simple FIFO queue implementations. CQ,
// which is based on channels, and SQ and SQU which are based on
// slices. All queues have fixed capacities (specified at creation
// time) and support the typical Push / Pop operations. SQ and SQU are
// only allowed to have capacities that are powers of 2. CQ and SQ are
// thread safe. SQU in not. Queues store elements of type "interface{}".
//
// You can generate queue implementations specialized to specific
// element data-types using the mkq.bash script. See Makefile for an
// example.
//
package queue
