#!/bin/sh
KAFKA_HOME="/home/hiro/Downloads/kafka_2.13-3.7.0"

$KAFKA_HOME/bin/kafka-topics.sh --bootstrap-server localhost:9092 --topic export_file --create --partitions 3 --replication-factor 1