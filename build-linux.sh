echo "*** Running make clean ***"
sudo make clean

echo "*** Running make compile for linux / amd64 ***"
sudo env GOOS=linux GOARCH=amd64 make compile

echo "*** Deleting the plugin_exec_linux_amd64 folder ***"
rm -rf couchbase_plugin_linux_amd64

echo "*** Creating the plugin_exec_linux_amd64 folder ***"
mkdir couchbase_plugin_linux_amd64

echo "*** Copying the release artifacts to plugin_exec_linux_amd64 folder ***"
cp nr-couchbase-plugin-config.yml.sample couchbase_plugin_linux_amd64/
cp nr-couchbase-plugin-definition.yml couchbase_plugin_linux_amd64/
cp -R ./bin couchbase_plugin_linux_amd64/bin
