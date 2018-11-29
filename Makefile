# MIT License
#
# Copyright (c) 2018 User not found (samuelngs)
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

NAME = dem

EXTENSIONS_DIR = extensions
EXTENSIONS = $(shell ls $(EXTENSIONS_DIR) | xargs -I* echo *.so)

BUILD_DIR = output
BUILD_EXTENSIONS_DIR = $(BUILD_DIR)/extensions

DEF_EXTENSIONS_DIR = $(HOME)/.config/dem/plugins

$(NAME): $(BUILD_DIR) $(BUILD_EXTENSIONS_DIR) $(EXTENSIONS)
	@echo "Building $(NAME)"
	@go build -a -installsuffix cgo -o $(BUILD_DIR)/$(NAME)

$(BUILD_DIR) $(BUILD_EXTENSIONS_DIR) $(DEF_EXTENSIONS_DIR):
	mkdir -p "$@"

%.so:
	@echo "Building extension $@"
	@go build -buildmode=plugin -o $(BUILD_EXTENSIONS_DIR)/$@ $(EXTENSIONS_DIR)/$*/main.go

install: $(NAME) $(DEF_EXTENSIONS_DIR)
	sudo cp -rf $(BUILD_DIR)/$(NAME) /usr/local/bin/$(NAME)
	sudo cp -rf $(BUILD_EXTENSIONS_DIR)/* $(DEF_EXTENSIONS_DIR)

clean:
	rm -rf $(BUILD_DIR)

.PHONY: install clean
