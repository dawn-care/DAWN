import Accordion from "../components/Accordion";
import LoginPic from "../assets/user_guide/Login.png"
import CarepartnerPairingPic from "../assets/user_guide/CarepartnerPairing.png"
import AccountSetupPic from "../assets/user_guide/AccountSetup.png"

export default function UserGuide() {
  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">
        StartUp Caregiver Manual
      </h1>

      <div className="border p-10 max-w-5xl w-full rounded-lg bg-white shadow space-y-6 text-gray-800 leading-relaxed">
        <p>
          Start-Up is designed to gently begin each morning for your carepartner by automatically playing meaningful videos, photos, messages, and helpful information. This daily “start-up routine” can help them feel calm, oriented, and connected to the people they love.
        </p>

        <p>
          This guide will walk you through setting everything up step-by-step.
        </p>

        {/* What Start-Up Does */}
        <Accordion title="What Start-Up Does" defaultOpen>
          <p>
            Start-Up allows you to upload special media, like videos, photos, audio, reminders, and daily messages, that will automatically play for your carepartner each morning.
          </p>

          <p className="mt-4">To make sure everything works properly each day:</p>
          <ul className="list-disc ml-6 mt-2 space-y-1">
            <li>Keep your carepartner’s device <b>plugged in and charging</b></li>
            <li>Leave the Start-Up app <b>open on their device</b> (even overnight)</li>
            <li>
              Ensure that the device does <b>not automatically go to sleep</b> after a certain amount of time has passed. You can find these preferences in the device settings.
            </li>
          </ul>

          <p className="mt-4">
            Then, in the morning, their personalized content will play <b>automatically!</b>
          </p>
        </Accordion>

        {/* Setting Up */}
        <Accordion title="Setting Up Start-Up">
          <h3 className="text-lg font-semibold mt-4">Step 1: Create Your Account</h3>
          <p className="mt-2">Here is how to <b>create an account</b> if you do not have one already:</p>
          <ul className="list-disc ml-6 mt-2 space-y-1">
            <li>Go to the Start-Up website.</li>
            <li>Click <b>Log In. </b></li>
            <li>Choose <b>Sign in with Google.</b></li>
            <li><b>Log in</b> using your Gmail account.</li>
          </ul>
          <img
            src={LoginPic}
            alt="Login screen"
            className="mt-3 ml-12 w-full max-w-3xl rounded-lg border shadow"
          />

          <p className="mt-4">Some important information:</p>
          <ul className="list-disc ml-6 mt-2 space-y-1">
            <li>Right now, Google sign-in is the only option available.</li>
            <li>If you don’t have a Gmail account, please create one first, then return to sign up.</li>
          </ul>

          <hr></hr>

          <h3 className="text-lg font-semibold mt-6">Step 2: Set Up Your Carepartner’s Device</h3>
          <ul className="list-disc ml-6 mt-2 space-y-1">
            <li><b>Download</b> the Start-Up application.</li>
            <li><b>Open</b> the app.</li>
            <li>Tap <b>“Pair Device.”</b></li>
            <li>You will see a <b>pairing code</b>
                <ul className="list-disc ml-6 mt-2 space-y-1">
                    <li>Note that this code expires after 10 minutes</li>
                </ul>
            </li>
          </ul>
          <img
            src={CarepartnerPairingPic}
            alt="Carepartner pairing screen"
            className="mt-3 ml-12 w-full max-w-3xl rounded-lg border shadow"
          />

          <hr></hr>

          <h3 className="text-lg font-semibold mt-6">Step 3: Pair the Devices</h3>
          <p className="mt-2">Once you’re signed in to the caregiver portal:</p>
          <ul className="list-disc ml-6 mt-2 space-y-1">
            <li>A webpage will appear asking you to <b>input a pairing code.</b></li>
            <li><b>Enter</b> the pairing code displayed on the care partner device.</li>
            <li><b>Read and accept</b> the Terms and Conditions.</li>
          </ul>

          <img
            src={AccountSetupPic}
            alt="Account setup screen"
            className="mt-3 ml-12 w-full max-w-3xl rounded-lg border shadow"
          />

          <p className="mt-4">On the carepartner’s device:</p>
          <ul className="list-disc ml-6 mt-2 space-y-1">
            <li>Tap the <b>Verify</b> button</li>
            <li>Read the <b>Terms and Conditions</b></li>
            <li>Tap <b>Accept</b></li>
          </ul>

          <p className="mt-4">
            Once this is done, the two devices are linked!
          </p>
        </Accordion>

        {/* Using Start-Up */}
        <Accordion title="Using Start-Up">
          <p>
            As a caregiver, you can create a gentle morning experience using the following modules.
          </p>

          <hr></hr>

          <h3 className="text-lg font-semibold mt-4">Media</h3>
          <p>You can upload:</p>
          <ul className="list-disc ml-6 mt-2 space-y-1">
            <li>Videos</li>
            <li>Photos</li>
            <li>Audio recordings</li>
            <li>Favorite songs</li>
          </ul>
          <p className="mt-2">Just upload files directly from your computer.</p>

          <hr></hr>

          <h3 className="text-lg font-semibold mt-4">Messages</h3>
          <p>
            Messages help orient your carepartner and remind them of important things, such as:
          </p>
          <ul className="list-disc ml-6 mt-2 space-y-1">
            <li>Where they are</li>
            <li>Where you are</li>
            <li>A personal message from you</li>
            <li>Events happening that day</li>
          </ul>
          <p className="mt-2">Just input information into the fields provided.</p>

          <hr></hr>

          <h3 className="text-lg font-semibold mt-4">Widgets</h3>
          <p>Widgets provide live information, like:</p>
          <ul className="list-disc ml-6 mt-2 space-y-1">
            <li>Weather updates</li>
            <li>News</li>
          </ul>
          <p className="mt-2">Just decide what type of widget you would like to add.</p>

          <hr></hr>

          <h3 className="text-lg font-semibold mt-4">Scheduling</h3>
          <p>
            In order to decide in what order your careparter sees their modules, you can use our calendar feature to:
          </p>
          <ul className="list-disc ml-6 mt-2 space-y-1">
            <li>Choose what content plays each morning</li>
            <li>Set the order things appear</li>
            <li>Pick which days certain items show</li>
          </ul>

          <p className="mt-4">
            This helps you build a comforting daily routine for your carepartner!
          </p>

          <p className="mt-4">
            Additionally, you can choose to schedule a module for a particular time of day. In the Calendar feature:
          </p>
          <ul className="list-disc ml-6 mt-2 space-y-1">
            <li>Select the checkbox “Timed”</li>
            <li>Input the time you would like the message to appear</li>
          </ul>
        </Accordion>

        {/* Recommendations */}
        <Accordion title="Our Recommendations">
          <h3 className="text-lg font-semibold mt-4">Default Media</h3>
          <p>
            A default piece of media is required. We strongly recommend making this a video! To improve recording quality, we recommend the following:
          </p>
          <ul className="list-disc ml-6 mt-2 space-y-1">
            <li>Sit in a well-lit room</li>
            <li>Speak clearly and slowly</li>
            <li>Use a calm, warm tone</li>
            <li>Look into the camera</li>
          </ul>

          <p className="mt-4">In terms of content, you may want to say things like:</p>
          <ul className="list-disc ml-6 mt-2 space-y-1">
            <li>“Hello! Good morning.”</li>
            <li>“You’re at home and you’re safe.”</li>
            <li>“You’re well taken care of.”</li>
            <li>“I love you very much.”</li>
            <li>“I’ll see you soon.”</li>
          </ul>

          <p className="mt-4">
            Additionally, don’t be afraid to have multiple of your carepartner’s loved ones present! For example, you might want to include your entire family in the video. This all helps create reassurance and familiarity at the start of each day.
          </p>

          <hr></hr>

          <h3 className="text-lg font-semibold mt-6">Ideas for Additional Content</h3>
          <p>To make mornings more meaningful, consider adding:</p>
          <ul className="list-disc ml-6 mt-2 space-y-1">
            <li>Favorite songs</li>
            <li>Family photos</li>
            <li>Short voice messages</li>
            <li>Daily plans or appointments</li>
            <li>Gentle reminders</li>
            <li>Encouraging words</li>
          </ul>

          <p className="mt-4">
            Giving them something to look forward to can bring comfort and joy.
          </p>
        </Accordion>

        {/* Final Thoughts */}
        <Accordion title="Final Thoughts">
          <p>Start-Up is designed to help your carepartner feel:</p>
          <ul className="list-disc ml-6 mt-2 space-y-1">
            <li>Safe</li>
            <li>Connected</li>
            <li>Loved</li>
            <li>Oriented to their day</li>
          </ul>

          <p className="mt-4">
            Even small messages, familiar faces, and favorite memories can make a big difference.
          </p>

          <p className="mt-4 font-medium">
            You’re not just setting up an app. You’re helping them begin each day with comfort and connection!
          </p>
        </Accordion>

      </div>
    </div>
  );
}