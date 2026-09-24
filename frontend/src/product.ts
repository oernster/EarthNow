// The product's name, as the page knows it. Its one home is internal/product on
// the Go side (the identity test holds every other shipped file to that); the
// page asks for it once through About and hands it down in this context, so the
// heading mark, the donate button and the guide all read the same answer.
import {createContext, useContext} from 'react'

export const ProductName = createContext('')

/** useProductName answers the product's name; empty until the Go side answers. */
export function useProductName(): string {
    return useContext(ProductName)
}
