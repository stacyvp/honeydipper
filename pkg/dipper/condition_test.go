// Copyright 2019 Honey Science Corporation
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, you can obtain one at http://mozilla.org/MPL/2.0/.

// +build !integration

package dipper

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompare(t *testing.T) {
	assert.True(t, Compare("test value1", "test value1"), "Simple string compare should match the same string")
	assert.False(t, Compare("test value1", "test value2"), "Simple string compare should find diff between strings")

	assert.True(t, Compare("test value1", regexp.MustCompile("test value[1-3]")), "RegExp should match string")
	assert.True(t, Compare("1test value21", regexp.MustCompile("test value[1-3]")), "RegExp should match string variation")
	assert.False(t, Compare("test value5", regexp.MustCompile("test value[1-3]")), "RegExp should fail matching wrong string")

	assert.True(t, Compare("test value1", []interface{}{"wef", "ksdfj", "test value1", "we"}), "Should find match in list")
	assert.False(t, Compare("test value1", []interface{}{"wef", "ksdfj", "test value9", "we"}), "Should find no match in list")
	assert.True(t, Compare("test value8", []interface{}{"wef", "ksdfj", regexp.MustCompile("test value[0-9]"), "we"}), "Should find matching regexp in list")
}

func TestCompareAllStr(t *testing.T) {
	assert.True(t, CompareAll("test value1", "test value1"), "Simple string compare should match the same string")
	assert.False(t, CompareAll("test value1", "test value2"), "Simple string compare should find diff between strings")

	assert.True(t, CompareAll("test value1", regexp.MustCompile("test value[1-3]")), "RegExp should match string")
	assert.True(t, CompareAll("1test value21", regexp.MustCompile("test value[1-3]")), "RegExp should match string variation")
	assert.False(t, CompareAll("test value5", regexp.MustCompile("test value[1-3]")), "RegExp should fail matching wrong string")

	assert.True(t, CompareAll("test value1", []interface{}{"wef", "ksdfj", "test value1", "we"}), "Should find match in list")
	assert.False(t, CompareAll("test value1", []interface{}{"wef", "ksdfj", "test value9", "we"}), "Should find no match in list")
	assert.True(t, CompareAll("test value8", []interface{}{"wef", "ksdfj", regexp.MustCompile("test value[0-9]"), "we"}), "Should find matching regexp in list")
}

func TestCompareAllList(t *testing.T) {
	assert.True(t, CompareAll([]interface{}{"dsf", "wrong", "test value1"}, "test value1"), "List match one string")
	assert.False(t, CompareAll([]interface{}{"dsf", "wrong", "test value1"}, "test value2"), "List match no string")

	assert.True(t, CompareAll([]interface{}{"dsf", "wrong", "test value1"}, regexp.MustCompile("test value[1-3]")), "List match regexp")
	assert.False(t, CompareAll([]interface{}{"dsf", "wrong", "test value9"}, regexp.MustCompile("test value[1-3]")), "List match no regexp")

	assert.True(t, CompareAll([]interface{}{"dsf", "wrong", "test value1"}, []interface{}{"test value1", "no value", "another"}), "List match one string in list")
	assert.False(t, CompareAll([]interface{}{"dsf", "wrong", "test value1"}, []interface{}{"test value9", "no value", "another"}), "List match no string in list")

	assert.True(t, CompareAll([]interface{}{"dsf", "wrong", "test value1"}, []interface{}{regexp.MustCompile("test value[1-3]"), "no value", "another"}), "List match a regexp in list")
	assert.False(t, CompareAll([]interface{}{"dsf", "wrong", "test value1"}, []interface{}{regexp.MustCompile("test value[2-3]"), "no value", "another"}), "List match no regexp in list")

	assert.True(t, CompareAll([]interface{}{"test 1", "test 2", "test 3"}, []interface{}{":all:", regexp.MustCompile("test [1-3]")}), "List should match all")
	assert.False(t, CompareAll([]interface{}{"test 1", "test 2", "test 4"}, []interface{}{":all:", regexp.MustCompile("test [1-3]")}), "List has exception in matching")
	assert.True(t, CompareAll([]interface{}{"test 1", "test 2", "test 4"}, []interface{}{":all:", []interface{}{regexp.MustCompile("test [1-3]"), "test 4"}}), "List match all condition")
}

func TestCompareAllMap(t *testing.T) {
	assert.True(t, CompareAll(map[string]interface{}{"key1": "val1", "key2": "val2"}, map[string]interface{}{"key1": "val1"}), "map key/value matches key/condition")
	assert.False(t, CompareAll(map[string]interface{}{"key1": "val1", "key2": "val2"}, map[string]interface{}{"key1": "val0"}), "map key/value mismatches key/condition")
	assert.True(t, CompareAll(map[string]interface{}{"key1": "val1", "key2": "val2"}, map[string]interface{}{"key1": "val1", "key2": "val2"}), "map all key/value matches key/condition")
	assert.False(t, CompareAll(map[string]interface{}{"key1": "val1", "key2": "val2"}, map[string]interface{}{"key1": "val1", "key2": "val0"}), "map some key/value matches key/condition")
	assert.True(t, CompareAll(map[string]interface{}{"key1": "val1", "key2": "val2"}, map[string]interface{}{":absent:": "key3"}), "map should have key3 absent")
	assert.True(t, CompareAll(map[string]interface{}{"key1": "val1", "key2": "val2"}, map[string]interface{}{":absent:": regexp.MustCompile("key3[1-3]")}), "map should have key3* absent")
	assert.False(t, CompareAll(map[string]interface{}{"key1": "val1", "key2": "val2", "key3": "val3"}, map[string]interface{}{":absent:": regexp.MustCompile("key3[1-3]*")}), "map should have key3* absent and fail")
	assert.False(t, CompareAll(map[string]interface{}{"key1": "val1", "key2": "val2", "key3": "val3"}, "invalid condition"), "fail with invalid condition for map value")
	assert.False(t, CompareAll(map[string]interface{}{"key1": "val1", "key2": "val2", "key3": "val3"}, map[string]interface{}{"key4": "111"}), "fail with condition that missing value")
}
