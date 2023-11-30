// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/solarlune/ldtkgo"
	"github.com/solarlune/resound/effects"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"github.com/tinne26/etxt"
)

//go:embed assets/*
var assets embed.FS

const sampleRate int = 44100 // assuming "normal" sample rate
var context *audio.Context

var imageCache map[string]*ebiten.Image

func init() {
	context = audio.NewContext(sampleRate)
	imageCache = make(map[string]*ebiten.Image)
}

// Frame is a single frame of an animation, usually a sub-image of a larger
// image containing several frames
type Frame struct {
	Duration int           `json:"duration"`
	Position FramePosition `json:"frame"`
}

// FramePosition represents the position of a frame, including the top-left
// coordinates and its dimensions (width and height)
type FramePosition struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// FrameTags contains tag data about frames to identify different parts of an
// animation, e.g. idle animation, jump animation frames etc.
type FrameTags struct {
	Name      string `json:"name"`
	From      int    `json:"from"`
	To        int    `json:"to"`
	Direction string `json:"direction"`
}

// Frames is a slice of frames used to create sprite animation
type Frames []Frame

// SpriteMeta contains sprite meta-data, basically everything except frame data
type SpriteMeta struct {
	ImageName string      `json:"image"`
	FrameTags []FrameTags `json:"frameTags"`
}

// SpriteSheet is the root-node of sprite data, it contains frames and meta data
// about them
type SpriteSheet struct {
	Sprite Frames     `json:"frames"`
	Meta   SpriteMeta `json:"meta"`
	Image  *ebiten.Image
}

// Load a sprite image and associated meta-data given a file name (without
// extension)
func loadSprite(name string) *SpriteSheet {
	name = path.Join("assets", "sprites", name)
	log.Printf("loading %s\n", name)

	file, err := assets.Open(name + ".json")
	if err != nil {
		log.Fatalf("error opening file %s: %v\n", name, err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	var ss SpriteSheet
	json.Unmarshal(data, &ss)
	if err != nil {
		log.Fatal(err)
	}

	ss.Image = loadImage(name + ".png")

	return &ss
}

func loadSpriteWithOSOverride(name string) *SpriteSheet {
	fname := name
	log.Printf("loading %s\n from OS", fname)

	file, err := os.Open(fname + ".json")
	if err != nil {
		log.Printf("error opening file from OS %s: %v\n", fname, err)
		log.Println("delegating opening file to internal assets")
		return loadSprite(name)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		log.Println("delegating opening file to internal assets")
		return loadSprite(name)
	}

	var ss SpriteSheet
	json.Unmarshal(data, &ss)
	if err != nil {
		log.Println(err)
		log.Println("delegating opening file to internal assets")
		return loadSprite(name)
	}

	ss.Image = loadImageWithOSOverride(fname)

	return &ss
}

// Convenience function to load entity/checkpoint PNGs from the maps folder
func loadEntityImage(name string) *ebiten.Image {
	name = path.Join("assets", "maps", name) + ".png"
	return loadImage(name)
}

// Load an image from embedded FS into an ebiten Image object
func loadImage(name string) *ebiten.Image {
	if imageCache[name] != nil {
		log.Printf("loading %s\n from cache", name)
		return imageCache[name]
	}

	log.Printf("loading %s\n", name)

	file, err := assets.Open(name)
	if err != nil {
		log.Fatalf("error opening file %s: %v\n", name, err)
	}
	defer file.Close()

	raw, err := png.Decode(file)
	if err != nil {
		log.Fatalf("error decoding file %s as PNG: %v\n", name, err)
	}

	if raw == nil {
		log.Fatalf("error empty data for sprite file %s\n", name)
	}

	imageCache[name] = ebiten.NewImageFromImage(raw)
	return imageCache[name]
}

func loadImageWithOSOverride(name string) *ebiten.Image {

	name = name + ".png"
	assetName := path.Join("assets", "sprites", name)

	log.Printf("loading %s\n from OS", name)

	file, err := os.Open(name)
	if err != nil {
		log.Printf("error opening file from OS %s: %v\n", name, err)
		log.Println("delegating opening file to internal assets")
		return loadImage(assetName)
	}
	defer file.Close()

	raw, err := png.Decode(file)
	if err != nil {
		log.Printf("error decoding file %s as PNG: %v\n", name, err)
		log.Println("delegating opening file to internal assets")
		return loadImage(assetName)
	}

	if raw == nil {
		log.Printf("error empty data for sprite file %s\n", name)
		log.Println("delegating opening file to internal assets")
		return loadImage(assetName)
	}

	return ebiten.NewImageFromImage(raw)
}

// Load an project from embedded FS into an LDtk Project object
func loadMaps(name string) *ldtkgo.Project {
	log.Printf("loading %s\n", name)

	file, err := assets.Open(name)
	if err != nil {
		log.Fatalf("error opening file %s: %v\n", name, err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("error reading from file %s: %v\n", name, err)
	}

	// Load the LDtk Project
	maps, err := ldtkgo.Read(data)
	if err != nil {
		log.Fatalf("error parsing file %s as LDtk Project: %v\n", name, err)
	}

	return maps
}

// SoundType is a unique identifier to reference sound by name
type SoundType uint8

const (
	backgroundMusic SoundType = iota
	sfxSplash
	sfxSubmerge
	sfxUnderwater
)

// Sound stores and plays all the sound variants for one single soundType
type Sound struct {
	Audio      []SoundData
	LastPlayed *audio.Player
	LastIndex  int
	Volume     float64
	lowpass    *effects.LowpassFilter
}

// AddSound adds one new sound to the soundType
func (s *Sound) AddSound(f string, sampleRate int, context *audio.Context, v ...int) {
	var filename string

	variants := 1
	if len(v) > 0 {
		variants = v[0]
	}

	for i := 0; i < variants; i++ {
		if variants == 1 {
			filename = f + ".ogg"
		} else {
			filename = f + "-" + strconv.Itoa(i+1) + ".ogg"
		}

		s.Audio = append(s.Audio, loadSoundFile(filename, sampleRate))
	}
}

// SetVolume sets the volume of the audio
func (s *Sound) SetVolume(v float64) {
	if v >= 0 && v <= 1 {
		s.Volume = v
	}
}

// Play plays the audio or a random one if there are more
func (s *Sound) Play() {
	length := len(s.Audio)
	index := 0

	if length == 0 {
		return
	} else if length > 1 {
		index = rand.Intn(length)
	}

	s.PlayVariant(index)
}

// PlayVariant plays the selected audio
func (s *Sound) PlayVariant(i int) {
	if i >= len(s.Audio) || i < 0 {
		return
	}

	sound, err := vorbis.DecodeWithoutResampling(bytes.NewReader(s.Audio[i]))
	if err != nil {
		log.Printf("error decoding sound as Vorbis: %v\n", err)
	}

	lowpass := effects.NewLowpassFilter(sound).SetStrength(0.85).SetActive(false)

	audioPlayer, err := audio.NewPlayer(context, lowpass)
	if err != nil {
		log.Printf("error making audio player: %v\n", err)
	}

	s.LastIndex = i
	s.LastPlayed = audioPlayer
	s.lowpass = lowpass
	audioPlayer.SetVolume(s.Volume)
	audioPlayer.Play()
}

// Pause pauses the audio being played
func (s *Sound) Pause() {
	if s.LastPlayed != nil {
		s.LastPlayed.Pause()
	}
}

// Resume resumes the last played audio
func (s *Sound) Resume() {
	if s.LastPlayed != nil {
		s.LastPlayed.Play()
	}
}

// Next plays the next audio from the list
func (s *Sound) PlayNext() {
	i := s.LastIndex + 1
	if i >= len(s.Audio) {
		i = 0
	}
	s.PlayVariant(i)
}

// IsPlaying returns if the sound is playing
func (s *Sound) IsPlaying() bool {
	return s.LastPlayed != nil && s.LastPlayed.IsPlaying()
}

// LowPass toggles the sound's low-pass filter
func (s *Sound) LowPass(on bool) {
	s.lowpass.SetActive(on)
	if on {
		s.SetVolume(0.7)
	} else {
		s.SetVolume(0.5)
	}
}

// Sounds is a slice of sounds
type Sounds []*Sound

func (s *Sound) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(s.Audio), func(i, j int) { s.Audio[i], s.Audio[j] = s.Audio[j], s.Audio[i] })
}

// MusicLoop is an audio player that infinitely loops back to its start
type MusicLoop struct {
	*audio.Player
	tween *gween.Tween
}

// FadeOut fades out the music smoothly to 0% volume
func (m *MusicLoop) FadeOut() {
	m.tween = gween.New(0.5, 0, 1*60, ease.InExpo)
}

// FadeIn fades in the music smoothly to 50% volume
func (m *MusicLoop) FadeIn() {
	m.tween = gween.New(0, 0.5, 2*60, ease.InExpo)
	m.Play()
}

// Update the music volume for fade effects
func (m *MusicLoop) Update() {
	if m.tween != nil {
		volume, done := m.tween.Update(1)
		m.SetVolume(float64(volume))
		if done {
			m.tween = nil
			if volume == 0 {
				m.Pause()
			}
		}
	}
}

// NewMusicPlayer loads a sound into an audio player that can be used to play it
// as an infinite loop of music without any additional setup required
func NewMusicPlayer(data SoundData) *MusicLoop {
	music, err := vorbis.DecodeWithoutResampling(bytes.NewReader(data))
	if err != nil {
		log.Printf("error decoding sound as Vorbis: %v\n", err)
	}

	musicLoop := audio.NewInfiniteLoop(music, music.Length())
	musicPlayer, err := audio.NewPlayer(context, musicLoop)
	if err != nil {
		log.Fatalf("error making music player: %v\n", err)
	}
	return &MusicLoop{musicPlayer, nil}
}

// SoundData is bytes returned from a sound file
type SoundData []byte

// Load an OGG Vorbis sound file with 44100 sample rate and return its stream
func loadSoundFile(name string, sampleRate int) SoundData {
	log.Printf("loading %s\n", name)

	file, err := assets.Open(name)
	if err != nil {
		log.Fatalf("error opening file %s: %v\n", name, err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	return data
}

func loadFont(name string) *etxt.Font {
	font, fname, err := etxt.ParseEmbedFontFrom(name, assets)
	if err != nil {
		log.Fatalf("error parsing font %s: %v", name, err)
	}

	log.Println("loaded font:", fname)
	return font
}
