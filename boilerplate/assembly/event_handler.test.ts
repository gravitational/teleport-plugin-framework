import { getFixture } from '../vendor/handle_event_test';
import { events, plugin } from '../vendor/teleport/teleport';
import { handleEvent } from "./index"

// Main test function
export function test(): void {
    // Get event from fixture #1
    const request = getFixture(1)

    // Send test event to the plugin
    const responseData = handleEvent(request)
    assert(responseData != null, "handleEvent returned null response")
    
    // Decode the response
    const response = plugin.HandleEventResponse.decode(responseData)

    // Ensure that user login has not been changed
    const event = response.Event
    assert(event.UserCreate != null, "Event has changed")

    // Ensure that login has not been changed
    const userCreateEvent = event.UserCreate as events.UserCreate
    assert(userCreateEvent.User.Login == "foo", "Login has changed")
}
