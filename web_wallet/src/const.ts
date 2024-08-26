export const HRP = 'morpheus'
export const COIN_SYMBOL = "RED"
export const DECIMAL_PLACES = 9
export const MAX_TRANSFER_FEE = 10000000n

let isDevMode = typeof window === 'undefined' || window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1'
export const DEVELOPMENT_MODE = isDevMode

export const FAUCET_HOST = DEVELOPMENT_MODE ? 'http://localhost:8765' : ""
export const API_HOST = DEVELOPMENT_MODE ? 'http://localhost:9650' : ""


import snapPkgJson from "sample-metamask-snap-for-hypersdk/package.json"
export const SNAP_ID = DEVELOPMENT_MODE ? 'local:http://localhost:8080' : `npm:${snapPkgJson.name}`
