import { defineStore } from 'pinia'

export const useBacktestStore = defineStore('backtest', {
  state: () => ({
    lastResult: null as any,
    lastForm: null as any,
  }),
  actions: {
    setResult(result: any) {
      this.lastResult = result
    },
    setForm(form: any) {
      this.lastForm = form
    },
  },
  persist: true,
})
