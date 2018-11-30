
#!/bin/sh -e

# Copied w/ love from the excellent hypnoglow/helm-s3

if [ -n "${HELM_PUSH_PLUGIN_NO_INSTALL_HOOK}" ]; then
    echo "Development mode: not downloading versioned release."
    exit 0
fi

version="$(curl -s https://api.github.com/repos/maorfr/helm-restore/releases/latest | awk '/\"tag_name\":/{gsub( /[,\"]/,"", $2); print $2}')"
echo "Downloading and installing helm-restore ${version} ..."

url=""
if [ "$(uname)" = "Darwin" ]; then
    url="https://github.com/maorfr/helm-restore/releases/download/${version}/helm-restore-macos-${version}.tgz"
elif [ "$(uname)" = "Linux" ] ; then
    url="https://github.com/maorfr/helm-restore/releases/download/${version}/helm-restore-linux-${version}.tgz"
else
    url="https://github.com/maorfr/helm-restore/releases/download/${version}/helm-restore-windows-${version}.tgz"
fi

echo $url

cd $HELM_PLUGIN_DIR
mkdir -p "bin"
mkdir -p "releases/${version}"

# Download with curl if possible.
if [ -x "$(which curl 2>/dev/null)" ]; then
    curl -sSL "${url}" -o "releases/${version}.tgz"
else
    wget -q "${url}" -O "releases/${version}.tgz"
fi
tar xzf "releases/${version}.tgz" -C "releases/${version}"
mv "releases/${version}/restore" "bin/restore" || \
    mv "releases/${version}/restore.exe" "bin/restore"
rm -rf releases