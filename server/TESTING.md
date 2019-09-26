Set up go modules:
````
sudo /snap/bin/go mod download
````

Add the testing device:

````
sudo ip link add wgtest type wireguard
````

Run the test as root:

````
sudo /snap/bin/go test -tags integration
````