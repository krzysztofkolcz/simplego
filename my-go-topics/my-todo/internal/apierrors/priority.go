package apierrors

// highPrio holds error mappings that are checked before the normal list.
// Add here any errors that must take absolute precedence regardless of
// how deeply nested they are in a wrapped error chain.
var highPrio = []APIErrors{}
