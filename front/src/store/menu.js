import { defineStore } from '@mpxjs/pinia'
import { createCartItem, DEFAULT_MERCHANT_ID, DEFAULT_STORE_ID, getAssetUrl, getCartId, getCartItems, getMenu, getSelectedStoreId, getStores, removeCartItem, saveSelectedStoreId, updateCartItem } from '../api/menu'

function distanceBetween(latitude, longitude, store) {
  const radians = degrees => degrees * Math.PI / 180
  const latitudeDelta = radians(store.latitude - latitude)
  const longitudeDelta = radians(store.longitude - longitude)
  const haversine = Math.sin(latitudeDelta / 2) ** 2 +
    Math.cos(radians(latitude)) * Math.cos(radians(store.latitude)) *
    Math.sin(longitudeDelta / 2) ** 2
  return 6371 * 2 * Math.atan2(Math.sqrt(haversine), Math.sqrt(1 - haversine))
}

export const useMenuStore = defineStore('menu', {
  state: () => ({
    merchantId: DEFAULT_MERCHANT_ID,
    storeId: DEFAULT_STORE_ID,
    stores: [],
    cartId: '',
    brandName: '知赋小程序',
    series: [],
    activeSeriesId: 'seasonal',
    selectedProductId: null,
    quantity: 1,
    cartItems: []
  }),
  getters: {
    activeStore(state) {
      return state.stores.find(store => store.storeId === state.storeId) || null
    },
    activeSeries(state) {
      return state.series.find(series => series.id === state.activeSeriesId) || state.series[0] || null
    },
    selectedProduct(state) {
      if (!state.selectedProductId) return null
      for (const series of state.series) {
        const product = series.products.find(item => item.id === state.selectedProductId)
        if (product) return product
      }
      return null
    },
    selectedPrice() {
      if (!this.selectedProduct) return 0
      return this.selectedProduct.price + this.selectedProduct.options.reduce((total, group) => {
        return total + group.values[group.selectedIndex].extraPrice
      }, 0)
    },
    selectedSummary() {
      if (!this.selectedProduct) return ''
      return this.selectedProduct.options.map(group => group.values[group.selectedIndex].label).join('，')
    },
    cartCount(state) {
      return state.cartItems.reduce((total, item) => total + item.quantity, 0)
    },
    cartTotal(state) {
      return state.cartItems.reduce((total, item) => total + item.unitPrice * item.quantity, 0)
    },
    cartLines(state) {
      return state.cartItems.map(item => {
        let image = ''
        for (const series of state.series) {
          const product = series.products.find(product => product.id === item.productId)
          if (product) {
            image = product.image
            break
          }
        }
        return {
          ...item,
          image,
          lineTotal: item.unitPrice * item.quantity
        }
      })
    }
  },
  actions: {
    async loadStores() {
      this.stores = await getStores(this.merchantId)
      if (this.stores.length === 0) throw new Error('No active stores are configured')
      const savedStoreId = getSelectedStoreId(this.merchantId)
      const savedStore = this.stores.find(store => store.storeId === savedStoreId)
      const selectedStore = savedStore || this.stores.find(store => store.isDefault) || this.stores[0]
      this.storeId = selectedStore.storeId
      saveSelectedStoreId(this.merchantId, this.storeId)
      return selectedStore
    },
    async selectNearestStore(location) {
      const locatedStores = this.stores.filter(store =>
        Number.isFinite(Number(store.latitude)) && Number.isFinite(Number(store.longitude)) &&
        (Number(store.latitude) !== 0 || Number(store.longitude) !== 0)
      )
      if (locatedStores.length === 0) return this.activeStore
      const nearestStore = locatedStores.reduce((nearest, store) => {
        return distanceBetween(location.latitude, location.longitude, store) < distanceBetween(location.latitude, location.longitude, nearest)
          ? store
          : nearest
      })
      if (nearestStore.storeId !== this.storeId) await this.selectStore(nearestStore.storeId)
      return nearestStore
    },
    async loadMenu() {
      this.cartId = getCartId(this.merchantId, this.storeId)
      const [menu, cartItems] = await Promise.all([
        getMenu(this.merchantId, this.storeId),
        getCartItems(this.merchantId, this.cartId, this.storeId)
      ])
      this.merchantId = menu.merchantId
      this.brandName = menu.brandName
      this.cartItems = cartItems
      this.series = menu.series.map(item => ({
        ...item,
        icon: getAssetUrl(item.icon),
        products: item.products.map(product => ({
          ...product,
          image: getAssetUrl(product.image)
        }))
      }))
      if (!this.series.some(item => item.id === this.activeSeriesId)) {
        this.activeSeriesId = this.series.length ? this.series[0].id : null
      }
      return menu
    },
    async selectStore(storeId) {
      const store = this.stores.find(item => item.storeId === storeId)
      if (!store) throw new Error('Selected store is unavailable')
      const previousStoreId = this.storeId
      this.storeId = store.storeId
      saveSelectedStoreId(this.merchantId, this.storeId)
      try {
        return await this.loadMenu()
      } catch (error) {
        this.storeId = previousStoreId
        this.cartId = getCartId(this.merchantId, previousStoreId)
        saveSelectedStoreId(this.merchantId, previousStoreId)
        throw error
      }
    },
    async refreshCart() {
      this.cartItems = await getCartItems(this.merchantId, this.cartId, this.storeId)
      return this.cartItems
    },
    selectSeries(seriesId) {
      this.activeSeriesId = seriesId
    },
    openProduct(productId) {
      this.selectedProductId = productId
      this.quantity = 1
    },
    closeProduct() {
      this.selectedProductId = null
    },
    selectOption({ groupIndex, optionIndex }) {
      if (!this.selectedProduct) return
      this.selectedProduct.options[groupIndex].selectedIndex = optionIndex
    },
    changeQuantity(amount) {
      this.quantity = Math.max(1, this.quantity + amount)
    },
    async addToCart() {
      if (!this.selectedProduct) return
      const item = await createCartItem(this.merchantId, this.cartId, {
        productId: this.selectedProduct.id,
        quantity: this.quantity,
        options: this.selectedProduct.options.map(group => ({
          id: group.id,
          selectedIndex: group.selectedIndex
        }))
      }, this.storeId)
      this.cartItems.push(item)
      this.closeProduct()
      return item
    },
    async setCartItemQuantity(itemId, quantity) {
      if (quantity < 1 || quantity > 99) return
      const item = await updateCartItem(this.merchantId, this.cartId, itemId, quantity, this.storeId)
      this.cartItems = this.cartItems.map(current => current.id === item.id ? item : current)
    },
    async removeCartItem(itemId) {
      await removeCartItem(this.merchantId, this.cartId, itemId, this.storeId)
      this.cartItems = this.cartItems.filter(item => item.id !== itemId)
    }
  }
})
