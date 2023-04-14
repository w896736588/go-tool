function setStore(key, value) {
  localStorage.setItem(key, value);
}

function getStore(key) {
  return localStorage.getItem(key);
}

export default {
  setStore,
  getStore,
}
