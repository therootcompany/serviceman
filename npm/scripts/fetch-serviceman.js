#!/usr/bin/env node

'use strict';
var os = require('os');

// https://nodejs.org/api/os.html#os_os_arch
// 'arm', 'arm64', 'ia32', 'mips', 'mipsel', 'ppc', 'ppc64', 's390', 's390x', 'x32', and 'x64'
var arch = os.arch(); // process.arch

// https://nodejs.org/api/os.html#os_os_platform
// 'aix', 'darwin', 'freebsd', 'linux', 'openbsd', 'sunos', 'win32'
var platform = os.platform(); // process.platform
var ext = 'windows' === platform ? '.exe' : '';

// This is _probably_ right. It's good enough for us
// https://github.com/nodejs/node/issues/13629
if ('arm' === arch) {
	arch += 'v' + process.config.variables.arm_version;
}

var map = {
	// arches
	armv6: 'armv6',
	armv7: 'armv7',
	arm64: 'armv8',
	ia32: '386',
	x32: '386',
	x64: 'amd64',
	// platforms
	darwin: 'darwin',
	linux: 'linux',
	win32: 'windows'
};

arch = map[arch];
platform = map[platform];

if (!arch || !platform) {
	console.error(
		"'" + os.platform() + "' on '" + os.arch() + "' isn't supported yet."
	);
	console.error(
		'Please open an issue at https://git.rootprojects.org/root/serviceman/issues'
	);
	process.exit(1);
}

var newVer = require('../package.json').version;
var fs = require('fs');
var exec = require('child_process').exec;
var request = require('@root/request');
var mkdirp = require('@root/mkdirp');

function needsUpdate(oldVer, newVer) {
	// "v1.0.0-pre" is BEHIND "v1.0.0"
	newVer = newVer.replace(/^v/, '').split(/[\.\-\+]/);
	oldVer = oldVer.replace(/^v/, '').split(/[\.\-\+]/);
	//console.log(oldVer, newVer);

	if (newVer[3] && !oldVer[3]) {
		// don't install beta over stable
		return false;
	}

	if (oldVer[3]) {
		if (oldVer[2] > 0) {
			oldVer[2] -= 1;
		} else if (oldVer[1] > 0) {
			oldVer[2] = 999;
			oldVer[1] -= 1;
		} else if (oldVer[0] > 0) {
			oldVer[2] = 999;
			oldVer[1] = 999;
			oldVer[0] -= 1;
		} else {
			// v0.0.0
			return true;
		}
	}
	if (newVer[3]) {
		if (newVer[2] > 0) {
			newVer[2] -= 1;
		} else if (newVer[1] > 0) {
			newVer[2] = 999;
			newVer[1] -= 1;
		} else if (newVer[0] > 0) {
			newVer[2] = 999;
			newVer[1] = 999;
			newVer[0] -= 1;
		} else {
			// v0.0.0
			return false;
		}
	}

	//console.log(oldVer, newVer);
	if (oldVer[0] > newVer[0]) {
		return false;
	} else if (oldVer[0] < newVer[0]) {
		return true;
	} else if (oldVer[1] > newVer[1]) {
		return false;
	} else if (oldVer[1] < newVer[1]) {
		return true;
	} else if (oldVer[2] > newVer[2]) {
		return false;
	} else if (oldVer[2] < newVer[2]) {
		return true;
	} else if (!oldVer[3] && newVer[3]) {
		return false;
	} else if (oldVer[3] && !newVer[3]) {
		return true;
	} else {
		return false;
	}
}

/*
// Same version
console.log(false === needsUpdate('0.5.0', '0.5.0'));
// No previous version
console.log(true === needsUpdate('', '0.5.1'));
// The new version is slightly newer
console.log(true === needsUpdate('0.5.0', '0.5.1'));
console.log(true === needsUpdate('0.4.999-pre1', '0.5.0-pre1'));
// The new version is slightly older
console.log(false === needsUpdate('0.5.0', '0.5.0-pre1'));
console.log(false === needsUpdate('0.5.1', '0.5.0'));
*/

exec('serviceman version', { windowsHide: true }, function(err, stdout) {
	var oldVer = (stdout || '').split(' ')[0];
	console.log(oldVer, newVer);
	if (!needsUpdate(oldVer, newVer)) {
		console.info(
			'Current serviceman version is new enough:',
			oldVer,
			newVer
		);
		return;
		//} else {
		//	console.info('Current serviceman version is older:', oldVer, newVer);
	}

	var url = 'https://rootprojects.org/serviceman/dist/{{ .Platform }}/{{ .Arch }}/serviceman{{ .Ext }}'
		.replace(/{{ .Version }}/g, newVer)
		.replace(/{{ .Platform }}/g, platform)
		.replace(/{{ .Arch }}/g, arch)
		.replace(/{{ .Ext }}/g, ext);

	console.info('Installing from', url);
	return request({ uri: url, encoding: null }, function(err, resp) {
		if (err) {
			console.error(err);
			return;
		}

		//console.log(resp.body.byteLength);
		//console.log(typeof resp.body);
		var serviceman = 'serviceman' + ext;
		return fs.writeFile(serviceman, resp.body, null, function(err) {
			if (err) {
				console.error(err);
				return;
			}
			fs.chmodSync(serviceman, parseInt('0755', 8));

			var path = require('path');
			var localdir = '/usr/local/bin';
			fs.rename(serviceman, path.join(localdir, serviceman), function(
				err
			) {
				if (err) {
					//console.error(err);
				}
				// ignore
			});

			var homedir = require('os').homedir();
			var bindir = path.join(homedir, '.local', 'bin');
			return mkdirp(bindir, function(err) {
				if (err) {
					console.error(err);
					return;
				}

				var localsrv = path.join(bindir, serviceman);
				return fs.writeFile(localsrv, resp.body, function(err) {
					if (err) {
						console.error(err);
						return;
					}
					fs.chmodSync(localsrv, parseInt('0755', 8));
					console.info('Wrote', serviceman, 'to', bindir);
				});
			});
		});
	});
});
