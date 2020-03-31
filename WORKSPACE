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
    commit = "b4a14686f0a98096416fe1b4cb848e384fb2b22b",
    importpath = "go.opencensus.io",
)

go_repository(
    name = "org_golang_google_grpc",
    commit = "24a4d6eb88bfde69ca4ef12fa00aef31059f74ec",
    importpath = "google.golang.org/grpc",
)

go_repository(
    name = "org_golang_google_api",
    commit = "214a0d2663ab0d218fc5fee8d44e96019beb1151",
    importpath = "google.golang.org/api",
)

go_repository(
    name = "org_golang_google_genproto",
    commit = "fa694d86fc64c7654a660f8908de4e879866748d",
    importpath = "google.golang.org/genproto",
)

go_repository(
    name = "com_github_googleapis_gax_go",
    commit = "bd5b16380fd03dc758d11cef74ba2e3bc8b0e8c2",
    importpath = "github.com/googleapis/gax-go",
)

go_repository(
    name = "com_google_cloud_go",
    commit = "f23c43891e43fa5323eb751293c177f0a4196b1a",
    importpath = "cloud.google.com/go",
)

go_repository(
    name = "org_golang_x_oauth2",
    commit = "0f29369cfe4552d0e4bcddc57cc75f4d7e672a33",
    importpath = "golang.org/x/oauth2",
)

go_repository(
    name = "org_golang_x_sync",
    commit = "112230192c580c3556b8cee6403af37a4fc5f28c",
    importpath = "golang.org/x/sync",
)

go_repository(
    name = "org_golang_x_net",
    commit = "74dc4d7220e7acc4e100824340f3e66577424772",
    importpath = "golang.org/x/net",
)

go_repository(
    name = "com_github_kylelemons_godebug",
    commit = "a435ca668a924cbe28b15c21c2f9d46ed72e6783",
    importpath = "github.com/kylelemons/godebug",
)

go_repository(
    name = "com_github_hashicorp_golang_lru",
    commit = "7f827b33c0f158ec5dfbba01bb0b14a4541fd81d",
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
