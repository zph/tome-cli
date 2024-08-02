#!/usr/bin/env bats

load 'test_helper/bats-support/load'
load 'test_helper/bats-support/load'

bats_require_minimum_version 1.5.0

script="$BATS_TEST_DIRNAME/wrapper.sh"
export PATH="bin:$PATH"
export TOME_DEBUG=1

@test "Test command-execute branch" {
  run $script command-execute arg1 arg2
  [ "${output}" = "tome-cli exec arg2" ]
}

@test "Test command-help branch" {
  run $script command-help arg1 arg2
  [ "${output}" = "tome-cli help arg2" ]
}

@test "Test command-complete branch when called without folder" {
  run -1 $script command-complete
}

@test "Test command-complete branch that it errors out" {
  run -1 $script command-complete fake-folder
}

@test "Test run branch" {
  run $script run arg1 arg2
  [ "${output}" = "tome-cli exec arg2" ]
}

@test "Test init branch (new tome)" {
  run $script init arg1
  echo $status
  echo $output
  [ "${output}" = "tome-cli init arg1" ]
}

@test "Test init branch (old tome)" {
  run $script init arg1 . arg3 arg4
  echo $status
  echo $output
  [ "${output}" = "tome-cli init arg3" ]
}

@test "Test default branch" {
  run $script arg1 arg2
  [ "${output}" = "tome-cli arg1 arg2" ]
}

