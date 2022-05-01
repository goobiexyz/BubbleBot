<img
     alt=""
     src="https://user-images.githubusercontent.com/86029592/166123196-80383346-91ad-4ef4-8e96-5395f90dff96.png"
     width="500" />

# BubbleBot
A modular Discord bot framework for the Go programming language, built upon [DiscordGo](https://github.com/bwmarrin/discordgo).

This package is very early in development, so expect frequent changes to the public API.

## What does it do?
By itself, not much. That's because BubbleBot is not designed to be a fully featured Discord bot right out of the box. Instead, it provides a platform for designing add-ons that give the bot its functionality, which are referred to as "toys" (because I thought that sounded cute). If enough people start making toys, theoretically you should be able to pick and choose what features you'd like your bot to have and customize BubbleBot to your specific needs, not dissimilar to how PostCSS lets you create your own customized CSS framework through the use of extensions.

### Complete feature set
- Simplifies the process of creating and running a bot
- Extensible add-ons API (toys)
- Easy-to-use Discord event manager (work in progress, only listens to MessageCreate events as of right now)
- Helpful logging functions

## Example
For an example of BubbleBot in action, take a look at my personal Discord bot, [GracieBot](https://github.com/gracieart/graciebot), which includes a toy for slash commands, which you may find useful if you are designing your own bot.
