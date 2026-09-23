package update

// releasePubKey verifica checksums.txt de releases nuevos. Vacío = releases
// sin firma (anteriores); en ese caso update avisa y sigue solo con sha256.
// El maintainer genera el par con `hush-sign gen` y firma cada release con
// `hush-sign sign checksums.txt`, subiendo el .sig al release.
const releasePubKey = "b1fe45a0ad9a21eb4f6d9e54f0b2806070cd3bf20cb94279a7cf735cb9012153"
