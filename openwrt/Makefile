include $(TOPDIR)/rules.mk
include $(INCLUDE_DIR)/package.mk

PKG_NAME:=mtg
PKG_VERSION:=1.0.0
PKG_RELEASE:=1

PKG_LICENSE:=MIT
PKG_MAINTAINER:=liaodiansm@gmail.com

define Package/mtg
  SECTION:=utils
  CATEGORY:=Utilities
  TITLE:=MTProto Proxy
endef

define Package/mtg/description
  A lightweight Golang application compiled outside SDK and packaged for OpenWrt.
endef

define Build/Prepare
	mkdir -p $(PKG_BUILD_DIR)
	cp $(TOPDIR)/package/mtg/mtg $(PKG_BUILD_DIR)/
	cp $(TOPDIR)/package/mtg/config.toml $(PKG_BUILD_DIR)/
	cp $(TOPDIR)/package/mtg/mtg.init $(PKG_BUILD_DIR)/
endef

define Build/Compile
    true
endef

define Package/mtg/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/mtg $(1)/usr/bin/mtg
    
	$(INSTALL_DIR) $(1)/etc/mtg
	$(INSTALL_CONF) $(PKG_BUILD_DIR)/config.toml $(1)/etc/mtg/config.toml
    
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/mtg.init $(1)/etc/init.d/mtg
endef

$(eval $(call BuildPackage,mtg))
