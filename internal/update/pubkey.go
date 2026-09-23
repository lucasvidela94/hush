package update

// releasePubKey verifica checksums.txt de releases nuevos. Vacío = releases
// sin firma (anteriores); en ese caso update avisa y sigue solo con sha256.
// El maintainer genera el par con `hush-sign gen` y firma cada release con
// `hush-sign sign checksums.txt`, subiendo el .sig al release.
const releasePubKey = ""
