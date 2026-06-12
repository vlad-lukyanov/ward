# Copyright 2026 Gentoo Authors
# Distributed under the terms of the GNU General Public License v2

EAPI=8

inherit go-module

DESCRIPTION="Ward - OpenRC service watchdog with automatic restart"
HOMEPAGE="https://github.com/vlad-lukyanov/ward"
SRC_URI="https://github.com/vlad-lukyanov/ward/archive/v${PV}.tar.gz -> ${P}.tar.gz"

LICENSE="MIT"
SLOT="0"
KEYWORDS="~amd64 ~arm64 ~x86"
IUSE=""

DEPEND="app-admin/openrc"
RDEPEND="${DEPEND}"

src_compile() {
	go build -o bin/ward ./cmd/ward/ || die
}

src_install() {
	dobin bin/ward

	insinto /etc/ward
	doins config.yaml

	newinitd "${FILESDIR}/${PN}.initd" "${PN}"

	keepdir /var/log
}
