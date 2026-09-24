package update

// releasePubKeys verifica checksums.txt de cada release. Con claves
// configuradas la firma es OBLIGATORIA: un release sin .sig o con firma
// inválida se rechaza. Para rotar: agregá la nueva clave al final, publicá un
// release firmado con la vieja (migra a todos), y recién después firmá solo
// con la nueva. Los binarios v0.4.1 (clave ff60…, privada perdida) no pueden
// verificar firmas nuevas: reinstalar por npm.
var releasePubKeys = []string{
	"b1fe45a0ad9a21eb4f6d9e54f0b2806070cd3bf20cb94279a7cf735cb9012153",
	"96c3f40308cefc60591b25ddc1615c6a9d1120ea14fa571eb3a0d97d9baf5abb",
}
