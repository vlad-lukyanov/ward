# Copyright 2026 Gentoo Authors
# Distributed under the terms of the GNU General Public License v2

EAPI=8

inherit git-r3

DESCRIPTION="WARD - OpenRC service watchdog with automatic restart"
HOMEPAGE="https://github.com/vlad-lukyanov/ward"
EGIT_REPO_URI="https://github.com/vlad-lukyanov/ward.git"
EGIT_BRANCH="main"

LICENSE="MIT"
SLOT="0"
KEYWORDS="~amd64 ~arm64 ~x86"
IUSE=""

DEPEND="sys-apps/openrc"
RDEPEND="${DEPEND}"

src_compile() {
	go build -o bin/ward ./cmd/ward/ || die
}

src_install() {
	dobin bin/ward

	insinto /etc/ward
	doins config.yaml

	newinitd "${FILESDIR}/${PN}.initd" "${PN}"

	insinto /etc/logrotate.d
	doins "${FILESDIR}/${PN}.logrotate"

	keepdir /var/log
}
