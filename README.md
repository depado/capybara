<h1 align="center">Capybara</h1>
<h2 align="center">

  [![forthebadge](https://forthebadge.com/images/badges/made-with-go.svg)](https://forthebadge.com)[![forthebadge](https://forthebadge.com/images/badges/built-with-love.svg)](https://forthebadge.com)[![forthebadge](https://forthebadge.com/images/badges/uses-badges.svg)](https://forthebadge.com)

  ![Go Version](https://img.shields.io/badge/Go%20Version-latest-brightgreen.svg)
  [![Go Report Card](https://goreportcard.com/badge/github.com/Depado/quokka)](https://goreportcard.com/report/github.com/depado/capybara)
  [![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/depado/capybara/blob/master/LICENSE)
  [![Say Thanks!](https://img.shields.io/badge/Say%20Thanks-!-1EAEDB.svg)](https://saythanks.io/to/Depado)
</h2>

<h2 align="center">Capybara</h2>
<h3 align="center">gRPC wrapper around bbolt</h3>

## Introduction

The goal of capybara is to provide a way to use bbolt in a centralized manner,
while maintaining the flexibility of bbolt and adding some features on top.

- gRPC API with Protobuf
- Distributed Lock 

The need for capybara was simple: Creating a very simple kv database that can
be accessed from multiple running services.

## gRPC

[gRPC](https://grpc.io/) is a great tool that allows to generate bindings for 
multiple languages at once. Making use of 
[protobuf](https://developers.google.com/protocol-buffers), it is one of the
most efficient RPC framework available. 

### Security

### Docker images versions
