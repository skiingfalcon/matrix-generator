# Coding Challenge #144 - Matrix Falling Text Generator

This challenge is to build your own version of the Matrix’s falling text.

JOHN CRICKETT  
OCT 31, 2026

Hi, this is John with this week’s Coding Challenge.

🙏 Thank you for being a subscriber, I’m honoured to have you as a reader. 🎉

If there is a Coding Challenge you’d like to see, please let me know by replying to this email📧

## Coding Challenge #144 - Matrix Falling Text Generator

This challenge is to build your own falling text generator that simulates the falling green text from The Matrix. We’re calling it **ccmatrix**.

If you have seen the film, you will remember the moment the screen fills with streams of glowing characters raining down in the dark. ccmatrix brings that to your terminal. It is a small program, but it is a lot of fun to build, and it is a great way to learn how a terminal really works: how to draw anywhere on the screen, how colours are produced, how to read key presses without waiting for the Enter key, and how to react when the window changes size.

The best part is that you get to watch your work come alive as soon as it runs.

## The Challenge - Building The Matrix Falling Text Generator

In this challenge you’re going to build a program that will take over the terminal, fill it with columns of random characters that fall at different speeds, and give the whole thing the glow of the film. It will also accept a handful of command line options, so you can change the speed, the colours and the characters it uses.

The core of this challenge is the same in any language, so pick whatever you are most interested in having fun with / learning.

### Step Zero

In this introductory step you’re going to set your environment up ready to begin developing and testing your solution.

Choose your programming language. Most languages have libraries for drawing on the terminal and reading keys, or you can talk to the terminal directly using escape codes. Either approach is fine, though you will learn more by trying the second.

### Step 1

In this step your goal is to switch the terminal into raw mode and draw a single falling character.

Put the terminal into raw mode, which lets you read key presses the moment they happen rather than waiting for the user to press Enter, and gives you full control over where things are drawn on the screen. Then draw a single character at the top of one column and move it down, one row at a time, with a short delay between each move.

The program should keep running until the user presses `q`. When it exits, it must restore the terminal to the state it was in before your program started, including the cursor, the colours and the contents of the screen as far as is practical.

**Testing:** Run your program with no options:

```
ccmatrix
```

You should see a single character appear at the top of the screen and fall, one row at a time, until it reaches the bottom. Press `q` and check that you are returned to a normal, usable prompt, with the cursor visible and typed characters echoed back. Try pressing other keys such as `a` or `x` and confirm that nothing happens and your program keeps running.

### Step 2

In this step your goal is to draw the trail behind the falling character.

Rather than a single character moving down the screen, draw a trail behind it: the leading character bright white, followed by a green trail a few rows long, and behind the trail the column left blank so the characters appear to fall through the dark. As the head moves down, the trail should move with it.

**Testing:**

```
ccmatrix
```

You should see a bright white character with a green trail falling away behind it, down a single column. Compare the length of the trail and how it fades with the real cmatrix running in another window. Press `q` and confirm the terminal is still restored correctly.

### Step 3

In this step your goal is to fill the screen with many falling trails.

Extend your single column so that every column on the screen has its own falling trail. Each should start at a random time and fall at its own speed, so the screen looks interesting and reflects the screenshot above, rather than uniform.

**Testing:**

```
ccmatrix
```

You should see the screen fill with falling green characters, each column with a bright white head. Press `q` and check that you are returned to a normal, usable prompt, with the cursor visible and typed characters echoed back. Try pressing other keys such as `a` or `x` and confirm that nothing happens and your program keeps running.

### Step 4

In this step your goal is to add screensaver mode with the `-s` flag.

When `-s` is passed, your program should exit as soon as the user presses any key at all, not just `q`. If your system has a mouse include that too. As before, the terminal must be fully restored on exit.

**Testing:**

```
ccmatrix -s
```

Let the screen fill, then press the space bar. The program should exit straight away and leave you with a working prompt. Repeat the test with a few other keys, such as a letter, a digit, the Enter key and finally the mouse. Then run without `-s` and check that those same keys are ignored, so the difference between the two modes is clear.

### Step 5

In this step your goal is to add control over the speed of the animation.

Support the `-u` flag, which sets the delay between updates as a number from 0 to 10. The default is 4, and 0 is the fastest. Your program should also let the user change the speed while it is running by pressing the keys 0 to 9, taking effect straight away without restarting the animation.

Think about what should happen when someone passes a value outside the range, or something that is not a number. A clear error message and a non-zero exit code is the right way to handle it.

**Testing:**

```
ccmatrix -u 0
ccmatrix -u 10
ccmatrix
```

The first run should be noticeably faster than the third (which uses the default of 4), and the second should be noticeably slower. While one of them is running, press `0`, then `9`, and watch the animation speed up and slow down. Finally check the error handling:

```
ccmatrix -u 11
ccmatrix -u fast
```

Both should report a problem and exit without drawing anything.

### Step 6

In this step your goal is to add colour options.

Support the `-C` flag, which sets the colour of the trails. It should accept one of green (the default), red, blue, white, yellow, cyan, magenta or black. The bright leading character should stay white. Also support the `-r` flag, which displays the trails in rainbow colours instead.

**Testing:**

```
ccmatrix -C red
ccmatrix -C cyan
ccmatrix -r
```

The first two runs should show trails in the requested colours, with white heads. The third should show trails made up of several different colours. Also try an unknown colour:

```
ccmatrix -C purple
```

This should report the list of colours that are supported and exit.

### Step 7

In this final step your goal is to add Japanese characters and handle resizing.

Support the `-c` flag, which draws half-width katakana characters instead of the default letters and numbers. This is the closest you will get to the look of the film, so it is a very satisfying one to see working.

Your program should also react when the terminal window is resized, either by making it larger or smaller. After a resize the display should redraw correctly to fill the new size, with no leftover fragments, no crashes and no columns drawn outside the window.

**Testing:**

```
ccmatrix -c
```

You should see katakana characters falling down the screen. If they appear as boxes or question marks, check that your terminal uses a font that supports them and that it is set to UTF-8. Combine it with other options to check they work together:

```
ccmatrix -c -C cyan -u 2
```

Then, while any of the above is running, drag the window larger and smaller. The animation should carry on and fill the new space each time. Do the same by changing the font size, or by using your window manager to maximise and restore the window.

## Going Further

Once you have the core tool working, here are some ideas to take it further:

- Add the `b` and `B` flags, which make some or all of the characters bold
- Add the `a` flag for asynchronous scrolling, where columns move at different rates
- Add a message mode, where the falling characters briefly resolve into a line of text of your choosing
- Support a configuration file so a user can store their favourite options
- Make the characters in the trail flicker and change as they fall, rather than staying fixed
- Add support for true colour terminals, with smooth fades along each trail
- Measure how much CPU your program uses, and see if you can make it lighter without changing how it looks
- Make it into a full blown screensaver
