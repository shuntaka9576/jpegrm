NAME := jpegrm
DIST := dist

.PHONY: build build-all package-windows installer clean

build:
	go build -o $(NAME) .

build-all: clean
	mkdir -p $(DIST)
	GOOS=darwin  GOARCH=arm64 go build -o $(DIST)/$(NAME)-darwin-arm64   .
	GOOS=darwin  GOARCH=amd64 go build -o $(DIST)/$(NAME)-darwin-amd64   .
	GOOS=windows GOARCH=amd64 go build -o $(DIST)/$(NAME)-windows-amd64.exe .
	GOOS=linux   GOARCH=amd64 go build -o $(DIST)/$(NAME)-linux-amd64   .

package-windows:
	GOOS=windows GOARCH=amd64 go build -o $(DIST)/jpegrm.exe .
	cp README-windows.txt $(DIST)/README-windows.txt
	cd $(DIST) && zip jpegrm-windows.zip jpegrm.exe README-windows.txt
	rm -f $(DIST)/jpegrm.exe $(DIST)/README-windows.txt

installer: package-windows
	@echo "==> Run on Windows: iscc installer.iss"
	@echo "==> Output: dist/jpegrm-setup.exe"

package-release:
ifndef VERSION
	$(error VERSION is required. Usage: make package-release VERSION=v0.0.3)
endif
	rm -rf $(DIST)/$(NAME)-$(VERSION)
	mkdir -p $(DIST)/$(NAME)-$(VERSION)
	gh release download $(VERSION) --repo shuntaka9576/jpegrm --pattern "jpegrm-setup.exe" --dir $(DIST)/$(NAME)-$(VERSION)
	cp README-windows.txt $(DIST)/$(NAME)-$(VERSION)/README-windows.txt
	cd $(DIST) && zip -r $(NAME)-$(VERSION).zip $(NAME)-$(VERSION)
	rm -rf $(DIST)/$(NAME)-$(VERSION)
	@echo "==> Created: $(DIST)/$(NAME)-$(VERSION).zip"

clean:
	rm -rf $(DIST) $(NAME)
