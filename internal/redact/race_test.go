//go:build race

package redact

// El detector de carreras vuelve todo varias veces más lento: con él activo
// medir tiempos no dice nada del rendimiento real.
const raceEnabled = true
