# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

http_archive(
    name = "io_bazel_rules_go",
    urls = ["https://github.com/bazelbuild/rules_go/releases/download/0.17.0/rules_go-0.17.0.tar.gz"],
    sha256 = "492c3ac68ed9dcf527a07e6a1b2dcbf199c6bf8b35517951467ac32e421c06c1",
)

http_archive(
    name = "bazel_gazelle",
    urls = ["https://github.com/bazelbuild/bazel-gazelle/releases/download/0.16.0/bazel-gazelle-0.16.0.tar.gz"],
    sha256 = "7949fc6cc17b5b191103e97481cf8889217263acf52e00b560683413af204fcb",
)

load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies", "go_register_toolchains")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

go_rules_dependencies()
go_register_toolchains()

gazelle_dependencies()

go_repository(
    name = "io_opencensus_go",
    commit = "57c09932883846047fd542903575671cb6b75070",
    importpath = "go.opencensus.io",
)

go_repository(
    name = "org_golang_google_grpc",
    commit = "a718efe0f40854b1f8bf60c04bfc9f8e8f6296db",
    importpath = "google.golang.org/grpc",
)

go_repository(
    name = "org_golang_google_api",
    commit = "77d02fa16783d31ba92e59fc0a888a45635f433e",
    importpath = "google.golang.org/api",
)

go_repository(
    name = "org_golang_google_genproto",
    commit = "8819c946db4494a2259bf100a377f51aa585d893",
    importpath = "google.golang.org/genproto",
)

go_repository(
    name = "com_github_googleapis_gax_go",
    commit = "ddfab93c3faef4935403ac75a7c11f0e731dc181",
    importpath = "github.com/googleapis/gax-go",
)

go_repository(
    name = "com_google_cloud_go",
    commit = "f23c43891e43fa5323eb751293c177f0a4196b1a",
    importpath = "cloud.google.com/go",
)

go_repository(
    name = "org_golang_x_oauth2",
    commit = "3e8b2be1363542a95c52ea0796d4a40dacfb5b95",
    importpath = "golang.org/x/oauth2",
)

go_repository(
    name = "org_golang_x_sync",
    commit = "37e7f081c4d4c64e13b10787722085407fe5d15f",
    importpath = "golang.org/x/sync",
)

go_repository(
    name = "org_golang_x_net",
    commit = "65e2d4e15006aab9813ff8769e768bbf4bb667a0",
    importpath = "golang.org/x/net",
)

go_repository(
    name = "com_github_kylelemons_godebug",
    commit = "d65d576e9348f5982d7f6d83682b694e731a45c6",
    importpath = "github.com/kylelemons/godebug",
)

go_repository(
    name = "com_github_hashicorp_golang_lru",
    commit = "20f1fb78b0740ba8c3cb143a61e86ba5c8669768",
    importpath = "github.com/hashicorp/golang-lru",
)

http_archive(
    name = "io_bazel_rules_docker",
    sha256 = "aed1c249d4ec8f703edddf35cbe9dfaca0b5f5ea6e4cd9e83e99f3b0d1136c3d",
    strip_prefix = "rules_docker-0.7.0",
    urls = ["https://github.com/bazelbuild/rules_docker/archive/v0.7.0.tar.gz"],
)

load("@io_bazel_rules_docker//go:image.bzl", _go_image_repos = "repositories")

_go_image_repos()

load("@io_bazel_rules_docker//repositories:repositories.bzl", container_repositories = "repositories")

container_repositories()

load("@io_bazel_rules_docker//container:container.bzl", "container_pull")

container_pull(
    name = "ubuntu",
    registry = "gcr.io",
    repository = "cloud-marketplace-containers/google/ubuntu16_04",
    digest = "sha256:c81e8f6bcbab8818fdbe2df6d367990ab55d85b4dab300931a53ba5d082f4296",
)
