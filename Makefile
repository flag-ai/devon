.PHONY: build-ui install-ui dev-ui clean-ui

FRONTEND_DIR := frontend
STATIC_DIR := src/devon/ui/static

install-ui:
	cd $(FRONTEND_DIR) && npm install

build-ui: install-ui
	cd $(FRONTEND_DIR) && npm run build

dev-ui:
	cd $(FRONTEND_DIR) && npm run dev

clean-ui:
	rm -rf $(STATIC_DIR)
