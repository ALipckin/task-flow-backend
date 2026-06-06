package domain

func NotificationRecipients(event TaskEvent) []int {
	seen := make(map[int]struct{}, len(event.ObserversIDs)+2)
	recipients := make([]int, 0, len(event.ObserversIDs)+2)

	add := func(id int) {
		if id == 0 {
			return
		}
		if _, exists := seen[id]; exists {
			return
		}
		seen[id] = struct{}{}
		recipients = append(recipients, id)
	}

	for _, id := range event.ObserversIDs {
		add(id)
	}
	add(event.PerformerID)
	add(event.CreatorID)

	return recipients
}
