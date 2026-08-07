import * as Google from 'expo-auth-session/providers/google';
import * as WebBrowser from 'expo-web-browser';

WebBrowser.maybeCompleteAuthSession();

// The iOS client ID is used for native Apple platform validation of the
// consent screen; the web client ID is what expo-auth-session's proxy flow
// actually authenticates against. Both must also be listed in the backend's
// GOOGLE_OAUTH_CLIENT_IDS so it accepts ID tokens issued for either.
const IOS_CLIENT_ID = '364131500940-c2f0gh3i3ks7ecs9sjsqr7tp9dkpdn3m.apps.googleusercontent.com';
const WEB_CLIENT_ID = '364131500940-oj2nd9rh6fpbd2bn4im81l2vrj8pa131.apps.googleusercontent.com';

export function useGoogleSignIn() {
  return Google.useIdTokenAuthRequest({
    iosClientId: IOS_CLIENT_ID,
    clientId: WEB_CLIENT_ID,
  });
}
