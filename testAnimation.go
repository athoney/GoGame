package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/athoney/finalGame/hscan"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/martinlindhe/inputbox"
)

const (
	screenWidth  = 640
	screenHeight = 480

	frameOX     = 0
	frameOY     = 0
	frameWidth  = 32
	frameHeight = 32
	frameNum    = 9
)

var (
	err              error
	gopher           *ebiten.Image
	background       *ebiten.Image
	cactusPic        *ebiten.Image
	flowerPic        *ebiten.Image
	blackBar         *ebiten.Image
	caps             *ebiten.Image
	num              *ebiten.Image
	nonAlphaNum      *ebiten.Image
	playerOne        player
	cactus           food
	flower           food
	foodPresent      string
	gopherX, gopherY float64
	password         string
	hash             string
	result           string
)

type Game struct {
	count int
}

type player struct {
	image       *ebiten.Image
	xPos, yPos  float64
	speed       float64
	size        float64
	direction   string
	passLen     int
	superpowers int
}

type food struct {
	image      *ebiten.Image
	xPos, yPos float64
}

func init() {
	gopher, _, err = ebitenutil.NewImageFromFile("goSprites.png")
	background, _, err = ebitenutil.NewImageFromFile("background.png")
	cactusPic, _, err = ebitenutil.NewImageFromFile("cactus.png")
	blackBar, _, err = ebitenutil.NewImageFromFile("black.png")
	flowerPic, _, err = ebitenutil.NewImageFromFile("bobbingFlower.png")
	caps, _, err = ebitenutil.NewImageFromFile("CapsHat.png")
	num, _, err = ebitenutil.NewImageFromFile("numberWand.png")
	nonAlphaNum, _, err = ebitenutil.NewImageFromFile("NonAlphaNum.png")

	if err != nil {
		log.Fatal(err)
	}

	playerOne = player{gopher, screenWidth / 2.0, screenHeight / 2.0, 4, 2, "down", 3, 0}
	cactus = food{cactusPic, 0, 0}
	flower = food{flowerPic, 0, 0}

	foodPresent = "none"
	gopherX = (screenWidth / 2.0) + 16
	gopherY = (screenHeight / 2.0) + 16
}

func (g *Game) Update() error {
	movePlayer()
	if foodPresent != "none" {
		collide()
	} else {
		if playerOne.superpowers < 3 {
			newFeature()
		}
		newFood()
	}
	//Update gopher center
	gopherX = (math.Mod(playerOne.xPos, 640)) + (16 * playerOne.size)
	gopherY = (math.Mod(playerOne.yPos, 480)) + (16 * playerOne.size)
	playerOne.xPos = math.Mod(playerOne.xPos, 640)
	playerOne.yPos = math.Mod(playerOne.yPos, 480)
	g.count++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	//Background
	backOp := &ebiten.DrawImageOptions{}
	backOp.GeoM.Translate(0, 0)
	screen.DrawImage(background, backOp)

	//black bar
	barOp := &ebiten.DrawImageOptions{}
	barOp.GeoM.Translate(0, 0)
	if password == "" {
		screen.DrawImage(blackBar.SubImage(image.Rect(0, 0, 100, 40)).(*ebiten.Image), barOp)
	} else {
		screen.DrawImage(blackBar.SubImage(image.Rect(0, 0, 260, 80)).(*ebiten.Image), barOp)
	}

	//Food Options
	if foodPresent == "cactus" {
		cactusOP := &ebiten.DrawImageOptions{}
		cactusOP.GeoM.Scale(2, 2)
		cactusOP.GeoM.Translate(cactus.xPos, cactus.yPos)
		screen.DrawImage(cactus.image, cactusOP)
	} else if foodPresent == "flower" {
		flowerOp := &ebiten.DrawImageOptions{}
		flowerOp.GeoM.Scale(2, 2)
		flowerOp.GeoM.Translate(-float64(frameWidth)/2, -float64(frameHeight)/2)
		flowerOp.GeoM.Translate(flower.xPos, flower.yPos)
		i := (g.count / 5) % frameNum
		sx, sy := frameOX+i*frameWidth, frameOY
		screen.DrawImage(flower.image.SubImage(image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image), flowerOp)
	}

	//Draw Gopher
	playerOp := &ebiten.DrawImageOptions{}
	playerOp.GeoM.Scale(playerOne.size, playerOne.size)
	playerOp.GeoM.Translate(playerOne.xPos, playerOne.yPos)
	if playerOne.direction == "Down" {
		screen.DrawImage(playerOne.image.SubImage(image.Rect(0, 0, 32, 32)).(*ebiten.Image), playerOp)
	} else if playerOne.direction == "Right" {
		screen.DrawImage(playerOne.image.SubImage(image.Rect(0, 32, 32, 64)).(*ebiten.Image), playerOp)
	} else if playerOne.direction == "Left" {
		screen.DrawImage(playerOne.image.SubImage(image.Rect(0, 64, 32, 96)).(*ebiten.Image), playerOp)
	} else {
		screen.DrawImage(playerOne.image.SubImage(image.Rect(32, 64, 64, 96)).(*ebiten.Image), playerOp)
	}

	//Tell user about their guess and/or display length and power info
	info := ""
	if password != "" {
		info = "Your password: " + password + "\nThe hash: " + hash + "\n"
		if result != "Not Found" {
			info += "Your password was found\n"
		} else {
			info += "Your password was not found!\n"
		}
	}
	ebitenutil.DebugPrint(screen, info+"Len: "+strconv.Itoa(playerOne.passLen)+" \nSuper Powers: "+strconv.Itoa(playerOne.superpowers))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("4010 Black Hat Go: Final Game")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}

// Move the player depending on which key is pressed
func movePlayer() {
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		playerOne.yPos -= playerOne.speed
		playerOne.direction = "Up"
		password = ""
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		playerOne.yPos += playerOne.speed
		playerOne.direction = "Down"
		password = ""
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		playerOne.xPos -= playerOne.speed
		playerOne.direction = "Left"
		password = ""
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		playerOne.xPos += playerOne.speed
		playerOne.direction = "Right"
		password = ""
	}
	if inpututil.IsKeyJustReleased(ebiten.KeySpace) {
		powers := ""
		if playerOne.superpowers == 0 {
			powers = "you have no powers"
		} else if playerOne.superpowers == 1 {
			powers = "numbers"
		} else if playerOne.superpowers == 2 {
			powers = "numbers and uppercase letters"
		} else {
			powers = "numbers, uppercase letters, and non-alphanumeric characters"
		}
		info := "Enter a password of length: " + strconv.Itoa(playerOne.passLen) + " Your powers include: " + powers
		got, ok := inputbox.InputBox("Password Strength Tester", info, "")
		if ok {
			password, hash, result = checkPass(got)
		} else {
			fmt.Println("No value entered")
		}
	}

	if playerOne.xPos < 0 {
		playerOne.xPos = 640 + playerOne.xPos
	}
	if playerOne.yPos < 0 {
		playerOne.yPos = 480 + playerOne.yPos
	}
}

func collide() {
	if foodPresent == "cactus" {
		if gopherX <= cactus.xPos+64 && gopherX > cactus.xPos && gopherY <= cactus.yPos+64 && gopherY > cactus.yPos {
			playerOne.size += 0.5
			playerOne.passLen += 1
			foodPresent = "none"
		}
	} else if foodPresent == "flower" {
		if gopherX <= flower.xPos+64 && gopherX > flower.xPos && gopherY <= flower.yPos+64 && gopherY > flower.yPos {
			playerOne.superpowers += 1
			foodPresent = "none"
			if playerOne.superpowers == 1 {
				playerOne.image = num
			} else if playerOne.superpowers == 2 {
				playerOne.image = caps
			} else if playerOne.superpowers == 3 {
				playerOne.image = nonAlphaNum
			}
		}
	}
}

func newFood() {
	var min int = int(playerOne.size) * 32
	if rand.Intn(500) == 0 {
		cactus.xPos = float64(rand.Intn(screenWidth-(63+min)) + min)
		cactus.yPos = float64(rand.Intn(screenHeight-(63+min)) + min)
		foodPresent = "cactus"
	}
}

func newFeature() {
	var min int = int(playerOne.size) * 32
	if rand.Intn(800) == 0 {
		flower.xPos = float64(rand.Intn(screenWidth-(63+min)) + min)
		flower.yPos = float64(rand.Intn(screenHeight-(63+min)) + min)
		foodPresent = "flower"
	}
}

func checkPass(pass string) (string, string, string) {
	//enforce password length
	len := len(pass)
	if len > playerOne.passLen {
		len = int(playerOne.passLen)
	}
	pass = pass[0:len]

	//enforce superpowers
	level0, _ := regexp.Compile("[^a-zA-Z]+")
	level1, _ := regexp.Compile("[^a-zA-Z0-9]+")

	if playerOne.superpowers == 0 {
		pass = strings.ToLower(level0.ReplaceAllString(pass, ""))
	} else if playerOne.superpowers == 1 {
		pass = strings.ToLower(level1.ReplaceAllString(pass, ""))
	} else if playerOne.superpowers == 2 {
		pass = level1.ReplaceAllString(pass, "")
	}
	//fmt.Printf("\nPassword: %s\n", pass)
	data := []byte(pass)
	hash := md5.Sum(data)
	//fmt.Print(hex.EncodeToString(hash[:]))

	result := hscan.GuessSingle(hex.EncodeToString(hash[:]), "wordlist2.txt")
	return pass, hex.EncodeToString(hash[:]), result
}
