package transaction_zether

const PAYLOAD_LIMIT = 1 + 144 // entire payload header is mandatorily encrypted
// sender position in ring representation in a byte, uptp 256 ring
// 144 byte payload  ( to implement specific functionality such as delivery of keys etc), user dependent encryption
const PAYLOAD0_LIMIT = 144 // 1 byte has been reserved for sender position in ring representation in a byte, uptp 256 ring
