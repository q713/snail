# MIT License
#
# Copyright (c) 2023 Jakob Görgen
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

all: debian-package

run:
	go run main.go

LINUX_BIN:=bin/snail-linux
NAME:=snail
VERSION:=0.0.1
REVISION:=0.0.1
ARCHITECTURE:=amd64
PACKAGE_DIR:=bin/$(NAME)_$(VERSION)-$(REVISION)_$(ARCHITECTURE)
PACKAGE_BIN_DIR:=$(PACKAGE_DIR)/usr/local/bin
DEBIAN_DIR:=$(PACKAGE_DIR)/DEBIAN
CONTROL_FILE:=$(DEBIAN_DIR)/control

release:
	@echo "Compiling for 64 bit windows, Mac and linux"
	GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.Version=v$(VERSION)'" -o $(LINUX_BIN) main.go
	GOOS=windows GOARCH=amd64 go build -ldflags="-X 'main.Version=v$(VERSION)'" -o bin/snail-windows main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'main.Version=v$(VERSION)'" -o bin/snail-darwin main.go
	@echo "finished building binaries"

debian-package: release
	@echo "Start creating a debian package"
	@echo $(PACKAGE_BIN_DIR)
	mkdir -p $(PACKAGE_BIN_DIR)
	cp $(LINUX_BIN) $(PACKAGE_BIN_DIR)/$(NAME)
	mkdir -p $(DEBIAN_DIR)
	touch $(CONTROL_FILE)
	echo "Package: $(NAME)" >> $(CONTROL_FILE)
	echo "Version: $(VERSION)" >> $(CONTROL_FILE)
	echo "Architecture: amd64" >> $(CONTROL_FILE)
	echo "Maintainer: Unknown" >> $(CONTROL_FILE)
	echo "Description: Little game to play" >> $(CONTROL_FILE)
	dpkg-deb --build --root-owner-group $(PACKAGE_DIR)
	#dpkg-deb -Zxz --build --root-owner-group $(PACKAGE_DIR)
	@echo "Finished creating debian package"


clean:
	rm -rf bin
	rm -rf snail

help:
	@echo "run 'make run' or just 'make' to build and run the game"
	@echo "run 'make release' to build a bin folder containing several build for different OSes"
	@echo "run 'make clean' for cleanup"