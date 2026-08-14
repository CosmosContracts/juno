#!/bin/sh
# Shared, network-free release helpers. Inputs are deliberately strict because
# this file is also used by the privileged publication job.

validate_version() {
	case ${1-} in
	v31.[0-9]*.[0-9]*) printf '%s' "$1" | grep -Eq '^v31\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?$' ;;
	*) return 1 ;;
	esac
}

validate_commit() {
	printf '%s' "${1-}" | grep -Eq '^[0-9a-f]{40}$'
}

validate_sha256() {
	printf '%s' "${1-}" | grep -Eq '^[0-9a-f]{64}$'
}

resolve_local_tag_commit() {
	repository=$1
	tag=$2
	validate_version "$tag"
	git -C "$repository" show-ref --verify --quiet "refs/tags/$tag" || {
		echo "requested release is not an exact Git tag: $tag" >&2
		return 1
	}
	commit=$(git -C "$repository" rev-parse --verify "refs/tags/$tag^{commit}") || return 1
	validate_commit "$commit" || return 1
	printf '%s\n' "$commit"
}

# Fetch a tag through a private validation ref and print its ref object and
# peeled commit. Optional expected values bind later checks to an earlier
# decision. Both ls-remote and fetch are checked so a move between them fails.
resolve_remote_tag_identity() {
	repository=$1 remote=$2 tag=$3 expected_oid=${4-} expected_commit=${5-}
	validate_version "$tag" || return 1
	line=$(git -C "$repository" ls-remote --exit-code "$remote" "refs/tags/$tag") || return 1
	[ "$(printf '%s\n' "$line" | wc -l | tr -d ' ')" = 1 ] || return 1
	remote_oid=${line%%[[:space:]]*}
	validate_commit "$remote_oid" || return 1
	[ -z "$expected_oid" ] || [ "$remote_oid" = "$expected_oid" ] || {
		echo "release tag object moved: expected $expected_oid, found $remote_oid" >&2
		return 1
	}
	validation_ref="refs/release-validation/$tag"
	git -C "$repository" update-ref -d "$validation_ref" >/dev/null 2>&1 || true
	git -C "$repository" fetch --no-tags "$remote" "+refs/tags/$tag:$validation_ref" || return 1
	fetched_oid=$(git -C "$repository" rev-parse --verify "$validation_ref") || return 1
	[ "$fetched_oid" = "$remote_oid" ] || {
		echo "release tag moved while validating" >&2
		return 1
	}
	commit=$(git -C "$repository" rev-parse --verify "$validation_ref^{commit}") || return 1
	validate_commit "$commit" || return 1
	[ -z "$expected_commit" ] || [ "$commit" = "$expected_commit" ] || {
		echo "release tag peeled commit moved: expected $expected_commit, found $commit" >&2
		return 1
	}
	printf '%s %s\n' "$remote_oid" "$commit"
}

require_absent_http_status() {
	status=$1
	identity=$2
	case "$status" in
	404) return 0 ;;
	200) echo "refusing replay: $identity already exists" >&2 ;;
	*) echo "existence guard failed closed for $identity (HTTP $status)" >&2 ;;
	esac
	return 1
}

package_binary() {
	binary=$1 arch=$2 version=$3 commit=$4 out=$5
	validate_version "$version"
	validate_commit "$commit"
	case "$arch" in amd64 | arm64) ;; *)
		echo "unsupported architecture: $arch" >&2
		return 1
		;;
	esac
	: "${SOURCE_DATE_EPOCH:?SOURCE_DATE_EPOCH is required}"

	stage=$(mktemp -d)
	trap 'rm -rf "$stage"' EXIT INT TERM
	binary_name="junod-linux-$arch"
	archive_name="juno-$version-linux-$arch.tar.gz"
	mkdir -p "$out"
	install -m 0755 "$binary" "$stage/$binary_name"
	printf '{"version":"%s","commit":"%s","goos":"linux","goarch":"%s","source_date_epoch":%s}\n' \
		"$version" "$commit" "$arch" "$SOURCE_DATE_EPOCH" >"$stage/BUILD-METADATA.json"
	cp "$stage/$binary_name" "$out/$binary_name"
	TZ=UTC tar --sort=name --format=ustar --owner=0 --group=0 --numeric-owner \
		--mtime="@$SOURCE_DATE_EPOCH" -C "$stage" -cf - BUILD-METADATA.json "$binary_name" |
		gzip -n -9 >"$out/$archive_name"
	rm -rf "$stage"
	trap - EXIT INT TERM
}

generate_checksums() {
	out=$1
	(cd "$out" && find . -maxdepth 1 -type f \
		! -name SHA256SUMS -printf '%f\n' |
		LC_ALL=C sort | xargs -r sha256sum) >"$out/SHA256SUMS"
}
