package main

import (
  "log"
  "os"
  "math"
  "bytes"
)

/**
 * Piece of code to burn useless data on a CD-R with the
 * sole purpose of painting patterns. It's called µ-engraving
 * because the bits on a CD are engraved at a ~1.5 µm resolution.
 *
 * This code should work the same way with CD-RW as CD-R.
 *
 * To burn with Mac OS X:
 *   mkdir out
 *   go run micro-engraving.go pie > out/a.wav
 *   drutil burn -noverify -nofs -audio -notest -noappendable -erase -eject out
 *
 * TODO:
 * - try data vs audio. Does one work better than the other?
 * - try different values for dark/light. Does contrast improve?
 * - take F3 re-ordering into account.
 * - make calibration easier/automatic.
 *
 * Links with useful technical or general information:
 * - http://www.ecma-international.org/publications/files/ECMA-ST/Ecma-130.pdf
 * - https://www.ecma-international.org/publications/files/ECMA-ST/ECMA-394.pdf
 * - https://www.ecma-international.org/publications/files/ECMA-ST/ECMA-395.pdf
 * - IEC 60908:1999
 *   https://github.com/suvozit/CD-Copy-protect
 * - IEC 908
 * - http://moais.imag.fr/membres/jean-louis.roch/perso_html/COURS/CSCI-506b-TRAIT-ERREURS/documents/IEC908.pdf
 * - http://www.informit.com/articles/article.aspx?p=1746162&seqNum=2
 * - http://www.laesieworks.com/digicom/Storage_CD.html
 * - http://www.instructables.com/id/Burning-visible-images-onto-CD-Rs-with-data-beta/
 * - https://superuser.com/questions/442040/how-to-paint-the-data-layer-of-a-cd-using-a-cd-drive
 * - https://github.com/happycube/ld-decode/blob/master/cd-decoder.py
 * - https://github.com/sidneycadot/Laser2Wav/tree/master/python
 * - https://en.wikipedia.org/wiki/Compact_disc#Physical_details
 * - The physics of the compact disc by John A Cope
 *   http://leung.uwaterloo.ca/CHEM/750/Lectures%202007/Lect%20Mat/physics%20of%20cd%20pe930103.pdf
 */

type Pattern string

const (
  Pitch Pattern = "pitch"
  Bands Pattern = "bands"
  Pie Pattern = "pie"

  Wav_header_size int = 44
  Sample_rate int = 44100
  Samples int = 1400
)

func main() {
  pattern := Pattern(os.Args[1])
  logger := log.New(os.Stderr, "", 0)

  logger.Printf("creating pattern: %s\n", pattern)

  buf := bytes.Buffer{}

  wav_header(&buf)
  if buf.Len() != Wav_header_size {
    logger.Printf("incorrect header length")
    os.Exit(-1)
  }

  switch pattern {
    case Pitch:
      pitch(&buf, 440)
    case Bands:
      bands(&buf, 8)
    case Pie:
      pie(&buf, 0.25)
    default:
      logger.Printf("unknown pattern")
      os.Exit(-1)
  }
  if buf.Len() != Sample_rate * Samples * 4 + Wav_header_size {
    logger.Printf("incorrect total bytes. Expecting %d, got %d\n",
      Sample_rate * Samples * 4 + Wav_header_size,
      buf.Len())
    os.Exit(-1)
  }
  buf.WriteTo(os.Stdout)
}

func wav_header(buf *bytes.Buffer) {
  len := Sample_rate * 4 * Samples
  buf.WriteString("RIFF")                // riff_tag
  write_int32(buf, Wav_header_size + len - 8) // riff_length
  buf.WriteString("WAVE")                // wave_tag
  buf.WriteString("fmt ")                // fmt_tag
  write_int32(buf, 16)                   // fmt_length
  write_int16(buf, 1)                    // audio_format
  write_int16(buf, 2)                    // num_channels
  write_int32(buf, Sample_rate)          // sample_rate
  write_int32(buf, 176400)               // byte_rate (44100 * 16 * 2 / 8)
  write_int16(buf, 4)                    // block_align (16 * 2 / 8)
  write_int16(buf, 16)                   // bits_per_sample
  buf.WriteString("data")                // data_tag
  write_int32(buf, len)                  // data_length
}

/**
 * Creates a wav file which plays a fixed pitch sound. Used for
 * testing purpose.
 */
func pitch(buf *bytes.Buffer, frequency float64) {
  for i:=0; i<Samples; i++ {
    for j:=0; j<Sample_rate; j++ {
      s := float64(j) / float64(Sample_rate) * 2 * math.Pi
      t := int(math.Sin(s * frequency) * 0x7fff)
      // left
      write_int16(buf, int(t))
      // right
      write_int16(buf, int(t))
    }
  }
}

/**
 * Draws concentric bands.
 */
func bands(buf *bytes.Buffer, bands int) {
  for i:=0; i<bands; i++ {
    for j:=0; j<Sample_rate * Samples/bands; j++ {
      if i % 2 == 0 {
        write_int16(buf, 0x4040)
        write_int16(buf, 0x4040)
      } else {
        write_int16(buf, 0x4545)
        write_int16(buf, 0x4545)
      }
    }
  }
}

/**
 * Draws a pie.
 */
func pie(buf *bytes.Buffer, width float64) {
  radius := 25.0  // in mm
  pitch := 0.00148 // distance between tracks, in mm

  linear_speed := 1300.0 // TODO: how to figure out the right value for this?
  byte_length := linear_speed / 176400

  for {
    // calculate number of bytes at the current radius
    circ := 2 * math.Pi * radius / byte_length
    for j:=0.0; j<4; j++ {
      for k:=int(circ / 4 * (j-1)); k<int(circ / 4 * j); k++ {
        if int(j) % 2 == 0 {
          buf.WriteByte(0x40)
        } else {
          buf.WriteByte(0x45)
        }
        if buf.Len() == Sample_rate * Samples * 4 + Wav_header_size {
          return
        }
      }
    }
    radius += pitch
  }
}

func write_int32(buf *bytes.Buffer, v int) {
  buf.WriteByte(byte(v & 0xff))
  buf.WriteByte(byte((v >> 8) & 0xff))
  buf.WriteByte(byte((v >> 16) & 0xff))
  buf.WriteByte(byte((v >> 24) & 0xff))
}

func write_int16(buf *bytes.Buffer, v int) {
  buf.WriteByte(byte(v & 0xff))
  buf.WriteByte(byte((v >> 8) & 0xff))
}
