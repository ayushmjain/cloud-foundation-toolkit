#! /bin/bash
# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e
set -u

BATS_ASSERT_VERSION=$1

cd /build
wget "https://github.com/jasonkarns/bats-assert-1/archive/v${BATS_ASSERT_VERSION}.zip"
unzip "v${BATS_ASSERT_VERSION}.zip"
cp -r "bats-assert-1-${BATS_ASSERT_VERSION}" /usr/local/bats-assert
rm -rf "v${BATS_ASSERT_VERSION}.zip" "bats-assert-1-${BATS_ASSERT_VERSION}"
