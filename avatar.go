package main

import (
	"errors"
)

// ErrNoAvatar is the error that is returned when the
// Avatar instance is unable to provide an avatar URL.
var ErrNoAvatarURL = errors.New("chat: Unable to get an avatar URL.")

// Avatar represents types capable of representing
// user profile pictures.
type Avatar interface {
	// GetAvatarURL gets the avatar URL for the specified client,
	// or returns an error if something goes wrong.
	// ErrNoAvatarURL is returned if the object is unable to get
	// a URL for the specified client.
	// GetAvatarURL(c *client) (string, error)

	GetAvatarURL(ChatUser) (string, error)
}

type TryAvatars []Avatar

func (a TryAvatars) GetAvatarURL(u ChatUser) (string, error) {
	for _, avatar := range a {
		if url, err := avatar.GetAvatarURL(u); err == nil {
			return url, nil
		}
	}
	return "", ErrNoAvatarURL
}

/*
	Mengambil profile picture dari akun social media

	GitHub, Facebook,
*/
type AuthAvatar struct{}

var UseAuthAvatar AuthAvatar

// func (AuthAvatar) GetAvatarURL(c *client) (string, error) {
// 	if url, ok := c.userData["avatar_url"]; ok {
// 		if urlStr, ok := url.(string); ok {
// 			return urlStr, nil
// 		}
// 	}
// 	return "", ErrNoAvatarURL
// }

func (AuthAvatar) GetAvatarURL(u ChatUser) (string, error) {
	url := u.AvatarURL()
	if len(url) == 0 {
		return "", ErrNoAvatarURL
	}
	return url, nil
}

/*
	Mengambil profile picture dari aku Gravatar, yang dikait
	dengan alamat email user pengguna
*/
type GravatarAvatar struct{}

var UseGravatar GravatarAvatar

// func (GravatarAvatar) GetAvatarURL(c *client) (string, error) {
// 	// if email, ok := c.userData["email"]; ok {
// 	// 	if emailStr, ok := email.(string); ok {
// 	// 		m := md5.New()
// 	// 		io.WriteString(m, strings.ToLower(emailStr))
// 	// 		return fmt.Sprintf("//www.gravatar.com/avatar/%x", m.Sum(nil)), nil
// 	// 	}
// 	// }
// 	// return "", ErrNoAvatarURL

// 	if userid, ok := c.userData["userid"]; ok {
// 		if useridStr, ok := userid.(string); ok {
// 			return "//www.gravatar.com/avatar/" + useridStr, nil
// 		}
// 	}

// 	return "", ErrNoAvatarURL
// }

func (GravatarAvatar) GetAvatarURL(u ChatUser) (string, error) {
	return "//www.gravatar.com/avatar/" + u.UniqueID(), nil
}

type FileSystemAvatar struct{}

var UseFileSystemAvatar FileSystemAvatar

// func (FileSystemAvatar) GetAvatarURL(c *client) (string, error) {
// 	if userid, ok := c.userData["userid"]; ok {
// 		if useridStr, ok := userid.(string); ok {
// 			return "/avatars/" + useridStr + ".jpg", nil
// 		}
// 	}
// 	return "", ErrNoAvatarURL
// }

// func (FileSystemAvatar) GetAvatarURL(c *client) (string, error) {
// 	if userid, ok := c.userData["userid"]; ok {
// 		if useridStr, ok := userid.(string); ok {
// 			files, err := ioutil.ReadDir("avatars")
// 			if err != nil {
// 				return "", ErrNoAvatarURL
// 			}
// 			for _, file := range files {
// 				if file.IsDir() {
// 					continue
// 				}
// 				if match, _ := path.Match(useridStr+"*", file.Name()); match {
// 					return "/avatars/" + file.Name(), nil
// 				}
// 			}
// 		}
// 	}
// 	return "", ErrNoAvatarURL
// }

func (FileSystemAvatar) GetAvatarURL(u ChatUser) (string, error) {
	if files, err := ioutil.ReadDir("avatars"); err == nil {
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if match, _ := path.Match(u.UniqueID()+"*", file.Name()); match {
				return "/avatars/" + file.Name(), nil
			}
		}
	}
	return "", ErrNoAvatarURL
}
