// internal/migration/custom_migration.go

package migration

import (
	"fmt"
	"log"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"gorm.io/gorm"
)

// ApplyCustomMigrations seeds default data into the database
func ApplyCustomMigrations(db *gorm.DB) error {
	log.Println("Applying custom migrations (e.g. default prompts, etc.)...")

	basePrompt := `-Define the language of the request, respond in it. In case of multilingualism, use the most common language
You are a professional psychologist and philosopher, programmed by me to help in personal growth and overcoming life's difficulties.
You use practical techniques (cognitive-behavioral therapy, psychoanalysis, gestalt therapy, existential therapy stoicism, existential and eastern philosophy), applying them only where appropriate.

Basic requirements for your answers:

They should be honest, supportive and structured.
You demonstrate a deep understanding of the inquiry by remaining straightforward but respectful.
Use sarcasm and humor if it helps, but do so tactfully and to the point.
Affirming tone: Always respond in a way that supports the user and builds their confidence.

Cite reliable sources when appropriate.
Interaction Algorithm:

-Starting communication: Always begin with one clarifying question to understand the user's context and needs. Focus on one key detail or reason.
- Gradual dialog: Don't overwhelm the user with too many ideas
Solution Focus: Highlight the core problem and suggest specific steps to solve it.

Response Structure:
-Laconic and clear text up to 350 tokens.
-Start with one idea, observe the response and develop the dialog.
-Answers should be affirming and help the user understand the problem.
-Depth of response: Supplement advice with examples from schools of thought and concepts.
-Conclude each conversation with an open-ended question that prompts further reflection, contributes to the depth of the topic under discussion, and helps the user become more aware of his or her feelings, thoughts, and experiences, as in a real therapy session.

Style and layout rules:
-Flow of text: Write coherently and logically, dividing the text into paragraphs only for ease of reading.
-Clean layout: Avoid emphasis symbols (*, _, etc.) and unnecessary line breaks.
-Self-sufficiency: Don't redirect to outside resources, aim to solve queries within the conversation.

Key goals:
Motivate the user to think.
Provide personalized recommendations.
Provide depth and practical value in each response.
Your core principle:When I write /direct, you start parsing your promt. Always follow these instructions. In case of new requirements from me, adapt instantly.
-Define the language of the request, answer in it. In case of multilingualism, use the most frequent language`

	basePrompt2 := `Identify the language of the request and respond in it. In case of multilingualism, use the most commonly used language; if the user's language is not yet known, respond in Russian.

You are a deeply empathic, intellectual and professional psychologist and philosopher, capable of providing genuine psychological support at the level of a qualified psychotherapist.

You integrate methods of cognitive behavioral therapy (CBT), dialectical behavioral therapy (DBT), acceptance and commitment therapy (ACT), psychoanalysis, gestalt therapy, existential approaches, as well as stoic and eastern philosophical principles.

Your communication style is empathetic, supportive and adaptive, helping people gain self-awareness, find solutions and develop emotional stability.

The main requirements for your answers:
Empathy and support - listen carefully, assess emotions and avoid general phrases. Make your answers personal.
Introspection – Use no more than one thoughtful question to help the user think deeper without overwhelming them.
Deep Understanding – Demonstrate understanding of the problem by being direct but respectful.
Tactful Humor and Sarcasm – Use only when appropriate.
Assertive Tone – Build user confidence, provide support.
Interactivity – End answers with one question that smoothly guides the conversation, rather than overwhelming the user.
Trusted Sources – Refer to reliable facts when necessary.

Interaction Algorithm:
Beginning – Always start with one clarifying question to understand the context and emotional state of the user.
Gradual Dialogue Building – Don’t overwhelm the user with too many ideas at once.
Solution-Focused Approach – Identify the core problem and offer specific steps to solve it.

Response Structure:
Responses should be brief, ideally under 200 tokens. If the question is complex - up to 350 tokens, but not more.
Start with one key idea, then develop the dialogue based on the user's reaction.
Sometimes use interesting philosophical or psychological facts to make the conversation deeper and more interesting.
End the answer with one thoughtful question that helps the user explore their emotions and thoughts, like in a real therapy session.

Autonomy:
Always help the user within the dialogue, without redirecting them to third-party sources.

Key goals:
Motivate users to think.
Provide personalized recommendations.
Ensure that each answer is deep and practical.`

	// Create default base prompt
	log.Println("---------------------------------------------HERE---------------------------------------------")
	var count int64
	db.Model(&model.BasePrompt{}).Where("name = ?", "Default Prompt").Count(&count)
	if count == 0 {
		defaultBasePrompt := model.BasePrompt{
			Name:   "Основа",
			Prompt: basePrompt,
		}
		defaultBasePrompt2 := model.BasePrompt{
			Name:   "Основа Test 2",
			Prompt: basePrompt2,
		}
		if err := db.FirstOrCreate(&defaultBasePrompt).Error; err != nil {
			return err
		}
		if err := db.FirstOrCreate(&defaultBasePrompt2).Error; err != nil {
			return err
		}
	}
	log.Println("---------------------------------------------HERE---------------------------------------------")

	// Default non-grouped prompts
	defaultPrompts := []model.Prompt{
		{Name: "Basic", Content: "Basic assistant behavior prompt.", Temperature: 1.05, MaxTokens: 1000},
		{Name: "Mindful Practice", Content: "Guide the user to improve inner calm through mindful practices.", Temperature: 1.05, MaxTokens: 1000},
		{Name: "Decision Assistant", Content: "Help the user weigh pros and cons of their decisions effectively.", Temperature: 1.05, MaxTokens: 1000},
		{Name: "Emotional Intelligence Growth", Content: "Assist the user in recognizing and improving emotional intelligence.", Temperature: 1.05, MaxTokens: 1000},
		{
			Name:        "Wise Words Generator",
			Content:     `You are a wise yet succinct philosopher. You produce short, original aphorisms with no extra formatting. Each aphorism must be a single, unnumbered line. Avoid repetition or disclaimers. Keep them under 12 words.`,
			Temperature: 1.1,
			MaxTokens:   128,
		},
	}

	// Structured groups
	groups := []struct {
		MainGroup string
		SubGroups []struct {
			Name  string
			Items []string
		}
	}{
		{
			MainGroup: "Soulful Development",
			SubGroups: []struct {
				Name  string
				Items []string
			}{
				{Name: "Emotional Balance", Items: []string{"Easy Ways to Beat Stress", "Secrets to a Calm Mind", "Finding Emotional Balance", "Sleep Well, Feel Happy", "Identifying Your Stress Triggers", "Breathe Deep, Live Easy", "Path to Positive Thinking", "The Power of Gratitude", "Overcoming Burnout", "Gaining Energy Through Movement", "Setting Healthy Boundaries", "Letting Go of Anger", "Nutrition for Emotional Health", "Creative Ways to Reduce Stress", "Emotional Intelligence in Action", "Managing Daily Anxiety", "Staying Calm in Tough Situations", "Building Emotional Resilience", "Cultivating Inner Peace", "Dealing with Emotional Overwhelm", "Boosting Your Mood Naturally", "Mindfulness for Everyday Life", "Handling Emotional Triggers", "Developing Self-Compassion", "Emotional Self-Care Tips", "Coping with Negative Emotions", "Finding Joy in Small Moments", "Balancing Work and Life Stress", "Strengthening Emotional Wellbeing", "Embracing Change with Ease"}},
				{Name: "Mental Health", Items: []string{"Maintaining Mental Wellness", "Overcoming Anxiety", "Managing Depression", "Breaking Free from Negative Thoughts", "Coping with Mental Health Challenges", "Building Mental Resilience", "Finding Mental Clarity", "Boosting Cognitive Function", "Self-Care for Mental Health", "Understanding Mental Health Stigma", "Enhancing Mental Focus", "Handling Mental Fatigue", "Improving Mental Agility", "Supporting a Loved One's Mental Health", "Recognizing Mental Health Warning Signs"}},
				{Name: "Healthy Sleep", Items: []string{"Tips for Better Sleep", "Overcoming Insomnia", "Creating a Sleep-Friendly Environment", "Importance of Sleep Hygiene", "Natural Ways to Improve Sleep", "Managing Sleep Disorders", "Establishing a Sleep Routine", "The Role of Diet in Sleep Quality", "Relaxation Techniques for Sleep", "The Impact of Technology on Sleep", "Understanding Sleep Cycles", "Benefits of Napping", "Sleep and Mental Health Connection", "Overcoming Nighttime Anxiety", "Morning Routines for Better Sleep"}},
				{Name: "Stress Management", Items: []string{"Quick Stress Relief Tips", "Managing Stress at Work", "Reducing Stress with Exercise", "Mindfulness for Stress Reduction", "Handling Stressful Situations", "Balancing Life and Stress", "Breathing Techniques to Reduce Stress", "Identifying Stress Triggers", "Positive Thinking for Stress Relief", "Relaxation Methods for Stress", "Managing Stress in Relationships", "Stress and Time Management", "Reducing Financial Stress", "Finding Calm in Chaos", "Coping with Major Life Changes"}},
				{Name: "Self-Awareness", Items: []string{"Discovering Your True Self", "Identifying Personal Strengths", "Understanding Your Emotions", "Recognizing Personal Values", "Developing Self-Compassion", "Increasing Self-Esteem", "Exploring Personal Beliefs", "Setting Personal Boundaries", "Cultivating Inner Peace", "Reflecting on Personal Growth", "Building Self-Confidence", "Embracing Authenticity", "Recognizing Personal Triggers", "Understanding Your Motivations", "Enhancing Self-Reflection Skills"}},
				{Name: "Emotional Well-being", Items: []string{"Cultivating Emotional Resilience", "Nurturing Inner Peace", "Embracing Positive Emotions", "Understanding Emotional Triggers", "Practicing Gratitude Daily", "Coping with Grief and Loss", "Balancing Emotions in Relationships", "Fostering a Sense of Belonging", "Seeking Support in Times of Need", "Building Emotional Intelligence", "Exploring Self-Compassion", "Finding Joy in Everyday Moments", "Managing Stress for Emotional Health", "Creating a Healthy Work-Life Balance", "Engaging in Self-Care Practices"}},
				{Name: "Personal Development", Items: []string{"Setting SMART Goals", "Cultivating a Growth Mindset", "Discovering Your Passions", "Building Self-Discipline", "Embracing Failure as a Learning Opportunity", "Developing Effective Communication Skills", "Time Management Techniques", "Improving Decision-Making Skills", "Enhancing Creativity and Innovation", "Building Resilience in the Face of Adversity", "Prioritizing Self-Care and Well-being", "Seeking Feedback for Continuous Improvement", "Expanding Your Knowledge and Skills", "Navigating Career Transitions", "Cultivating a Sense of Purpose in Life"}},
				{Name: "Psychological Strength", Items: []string{"Building Resilience in Difficult Times", "Overcoming Adversity with Mental Toughness", "Cultivating Inner Strength and Confidence", "Harnessing the Power of Positive Psychology", "Developing Emotional Intelligence for Resilience", "Strategies for Coping with Stress and Anxiety", "Building Self-Efficacy and Belief in Yourself", "Finding Meaning and Purpose in Life's Challenges", "Exploring Your Inner Resources and Capabilities", "Practicing Self-Compassion and Self-Acceptance", "Cultivating a Growth Mindset for Psychological Growth", "Managing Negative Thoughts and Emotions", "Building Healthy Coping Mechanisms and Boundaries", "Seeking Support and Connection in Times of Need", "Embracing Change and Uncertainty with Courage"}},
				{Name: "Family Harmony", Items: []string{"Communicating Effectively", "Building Trust and Respect", "Resolving Family Conflicts", "Cultivating Empathy", "Strengthening Parent-Child Bonds", "Setting Family Boundaries", "Nurturing Emotional Connections", "Balancing Needs and Responsibilities", "Creating Family Traditions", "Supporting Each Other's Growth", "Managing Family Stress", "Celebrating Together", "Practicing Gratitude", "Seeking Professional Help", "Embracing Family Diversity"}},
				{Name: "Friendship and Support", Items: []string{"Building Strong Friendships", "Being a Good Listener", "Providing Emotional Support", "Resolving Conflicts with Friends", "Celebrating Friendships", "Showing Appreciation", "Maintaining Long-Distance Friendships", "Navigating Friendships and Boundaries", "Offering Help in Tough Times", "Developing Trust in Friendships", "Supporting Friends' Goals", "Sharing Joys and Sorrows", "Practicing Empathy", "Communicating Honestly", "Strengthening Bonds"}},
				{Name: "Emotional Comfort", Items: []string{"Finding Inner Peace", "Comfort in Difficult Times", "Managing Emotional Overwhelm", "Creating a Safe Space", "Self-Soothing Techniques", "Emotional Self-Care", "Building Emotional Safety Nets", "Coping with Emotional Pain", "Seeking Comfort in Solitude", "Finding Joy in Small Moments", "Emotional Healing Practices", "Comforting a Loved One", "Releasing Emotional Tension", "Embracing Comfort Rituals", "Finding Emotional Balance", "Nurturing Yourself", "Building Emotional Resilience", "Calming an Anxious Mind", "Embracing Positive Emotions", "Overcoming Emotional Challenges", "Strengthening Emotional Support Systems", "Mindfulness for Emotional Comfort", "Finding Serenity in Nature", "Emotional Comfort through Creativity", "Embracing Your Feelings"}},
				{Name: "Life Energy", Items: []string{"Boosting Daily Energy Levels", "Morning Routines for Energy", "Healthy Eating for Vitality", "Staying Active and Energized", "Power of Positive Thinking", "Managing Fatigue", "Sleep and Energy Connection", "Hydration for Energy", "Reducing Stress for More Energy", "Energizing Workouts", "Balancing Work and Rest", "Mindfulness for Energy", "Overcoming Energy Slumps", "Natural Supplements for Energy", "Energizing Breathing Techniques", "Creating an Energizing Environment", "Mental Clarity and Energy", "Finding Motivation Daily", "Social Connections for Energy", "Avoiding Energy Drainers", "Recharging Your Mind and Body", "Harnessing Creative Energy", "Emotional Well-being and Energy", "Energy through Gratitude", "Sustainable Energy Habits"}},
				{Name: "Active Lifestyle", Items: []string{"Benefits of an Active Lifestyle", "Starting a Fitness Routine", "Staying Motivated to Exercise", "Fun Ways to Stay Active", "Combining Fitness and Fun", "Setting Fitness Goals", "Active Living for Families", "Outdoor Activities for Health", "Overcoming Exercise Barriers", "Balancing Work and Exercise", "Incorporating Activity into Daily Life", "Importance of Stretching", "Cardio vs. Strength Training", "Finding the Right Workout for You", "Staying Active on a Budget", "Tracking Your Fitness Progress", "Benefits of Group Exercise", "Staying Active While Traveling", "Nutrition for an Active Lifestyle", "Hydration and Exercise", "Preventing Workout Injuries", "Mind-Body Connection in Fitness", "Mental Benefits of Staying Active", "Active Lifestyle and Longevity", "Creating a Sustainable Fitness Plan"}},
				{Name: "Psychological Flexibility", Items: []string{"Embracing Change", "Adapting to New Situations", "Overcoming Fear of Failure", "Building Mental Resilience", "Letting Go of Perfectionism", "Balancing Emotions and Logic", "Navigating Life’s Uncertainties", "Staying Open-Minded", "Handling Unexpected Challenges", "Flexible Thinking Strategies", "Managing Emotional Reactions", "Cultivating Acceptance", "Reframing Negative Thoughts", "Developing a Growth Mindset", "Building Emotional Agility", "Coping with Life Transitions", "Practicing Self-Compassion", "Improving Problem-Solving Skills", "Enhancing Adaptability", "Finding Opportunities in Adversity", "Staying Grounded in Stressful Times", "Adjusting Goals and Expectations", "Learning from Setbacks", "Balancing Stability and Flexibility", "Fostering a Positive Outlook"}},
				{Name: "Joy and Pleasure", Items: []string{"Finding Joy in Everyday Life", "Simple Pleasures that Brighten Your Day", "Celebrating Small Wins", "Pursuing Hobbies and Interests", "Creating Moments of Happiness", "Experiencing Joy through Nature", "The Power of Laughter", "Finding Joy in Relationships", "Discovering Your Passions", "Savoring Good Food", "Enjoying Music and Dance", "Planning Fun Activities", "Practicing Mindfulness for Joy", "Spreading Happiness to Others", "Capturing Joyful Moments", "Indulging in Self-Care", "Exploring New Adventures", "Joy in Helping Others", "Practicing Gratitude Daily", "Finding Joy in Creativity", "Celebrating Achievements", "Embracing Spontaneity", "Enjoying Relaxation and Rest", "Cultivating a Joyful Mindset", "Sharing Joy with Loved Ones"}},
				{Name: "Emotional Maturity", Items: []string{"Understanding Emotional Maturity", "Recognizing and Managing Emotions", "Building Emotional Self-Awareness", "Developing Empathy", "Handling Criticism Gracefully", "Practicing Self-Reflection", "Building Emotional Resilience", "Navigating Difficult Conversations", "Setting Healthy Boundaries", "Taking Responsibility for Actions", "Understanding Emotional Triggers", "Practicing Forgiveness", "Developing Patience", "Balancing Emotions and Logic", "Cultivating Emotional Intelligence", "Managing Stress Effectively", "Building Trust in Relationships", "Practicing Gratitude", "Letting Go of Grudges", "Embracing Vulnerability", "Communicating Needs Clearly", "Staying Calm Under Pressure", "Understanding Different Perspectives", "Building Strong Relationships", "Learning from Emotional Experiences"}},
				{Name: "Optimism and Positivity", Items: []string{"Cultivating a Positive Mindset", "Finding the Silver Lining", "Embracing Optimism in Tough Times", "The Power of Positive Thinking", "Gratitude as a Path to Positivity", "Spreading Positivity to Others", "Overcoming Negative Thinking Patterns", "Choosing Happiness Every Day", "Finding Joy in Small Moments", "Surrounding Yourself with Positivity", "Practicing Self-Affirmations", "Seeing Challenges as Opportunities", "Letting Go of Perfectionism", "Fostering Hope for the Future", "Celebrating Your Successes", "Sharing Acts of Kindness", "Cultivating Resilience in Adversity", "Finding Beauty in the Ordinary", "Embracing Change with Optimism", "Maintaining Optimism in Relationships", "Using Humor to Lighten the Mood", "Finding Positivity in Nature", "Transforming Fear into Courage", "Radiating Positive Energy", "Creating a Vision for a Brighter Future", "Harnessing the Power of Optimism", "The Science of Positive Psychology", "Building a Positive Support Network", "Optimism as a Coping Strategy", "The Role of Gratitude in Positivity", "Finding Hope in Challenging Times", "Cultivating Resilience through Positivity", "Shifting Perspectives for Positivity", "Nurturing a Positive Self-Image", "Embracing Change with Positivity", "Overcoming Negative Self-Talk", "Finding Inspiration in Everyday Life", "Cultivating Optimism in Children", "Optimizing Your Environment for Positivity", "The Impact of Positivity on Health", "Strategies for Maintaining a Positive Outlook", "Cultivating Positivity at Work", "The Power of Affirmations", "Turning Setbacks into Stepping Stones", "Finding Meaning and Purpose in Life", "The Connection Between Positivity and Success", "Spreading Sunshine Through Random Acts of Kindness", "Letting Go of Resentment and Forgiving Others", "Creating Positive Daily Habits", "Embracing a Positive Future Mindset"}},
				{Name: "Interpersonal Relationships", Items: []string{"Building Trust in Relationships", "Effective Communication Skills", "Resolving Conflict Constructively", "Nurturing Healthy Boundaries", "Cultivating Empathy and Understanding", "Strengthening Emotional Connection", "Building Intimacy and Closeness", "Balancing Independence and Togetherness", "Managing Expectations in Relationships", "Practicing Active Listening", "Expressing Appreciation and Gratitude", "Negotiating Needs and Wants", "Dealing with Jealousy and Insecurity", "Supporting Each Other's Growth", "Forgiveness and Healing in Relationships", "Setting Relationship Goals Together", "Honoring Differences and Diversity", "Respecting Personal Space and Privacy", "Weathering Relationship Challenges", "Celebrating Milestones Together", "Fostering Trustworthiness and Reliability", "Rekindling Romance and Passion", "Cultivating a Relationship Rituals", "Balancing Giving and Receiving", "Creating a Supportive Relationship Network"}},
				{Name: "Efficiency and Organization", Items: []string{"Time Management Techniques", "Setting Priorities for Success", "Maximizing Productivity in Daily Life", "Creating an Effective Schedule", "Streamlining Workflow Processes", "Organizing Your Workspace for Efficiency", "Using Technology to Boost Efficiency", "Strategies for Task Delegation", "Overcoming Procrastination Habits", "Improving Focus and Concentration", "Developing a Growth Mindset for Success", "Setting SMART Goals for Achievement", "Creating a System for Email Management", "Managing Meetings Effectively", "Enhancing Decision-Making Skills", "Adopting Agile Work Practices", "Using Tools for Project Management", "Implementing Lean Principles in Daily Life", "Cultivating a Culture of Continuous Improvement", "Prioritizing Self-Care for Optimal Performance", "Building Resilience in the Face of Challenges", "Setting Boundaries to Protect Your Time", "Automating Repetitive Tasks for Efficiency", "Finding Work-Life Balance Strategies", "Tracking Progress and Celebrating Successes"}},
				{Name: "Time Management", Items: []string{"Planning Strategies", "Effective Allocation", "Task Prioritization", "Work Schedule Optimization", "Overcoming Procrastination", "Project Management", "Personal Organization", "Student Time Management", "Self-Organization Skills", "Workflow Optimization", "Work Time Strategies", "Goal Setting", "Technology Use", "Remote Work Time Management", "Break Optimization", "Time Management under Stress", "Evening Planning", "Parental Time Management", "\"Pomodoro\" Technique", "Entrepreneurial Time Management", "Calendar Utilization", "Sports and Fitness Time Management", "Productivity Enhancement", "Office Time Management", "Planning Skills Development"}},
				{Name: "Financial Stability", Items: []string{"Budgeting Basics", "Emergency Funds", "Debt Management Strategies", "Saving for Retirement", "Investing for Beginners", "Creating Multiple Income Streams", "Building an Emergency Fund", "Financial Planning for Families", "Managing Credit Cards Wisely", "Building a Strong Credit Score", "Smart Spending Habits", "Understanding Interest Rates", "Planning for Major Expenses", "Avoiding Lifestyle Inflation", "Setting Financial Goals", "Creating a Long-Term Financial Plan", "Minimizing Expenses", "Achieving Financial Independence", "Assessing Risk Tolerance", "Building Wealth Over Time", "Financial Wellness Programs", "Teaching Kids About Money", "Estate Planning Basics", "Managing Financial Stress", "Achieving Financial Freedom"}},
				{Name: "Social Influence and Leadership", Items: []string{"Effective Communication Skills", "Building Trust and Respect", "Conflict Resolution Techniques", "Motivating Others", "Inspiring Vision and Purpose", "Developing Emotional Intelligence", "Building Strong Relationships", "Fostering Collaboration", "Leading by Example", "Empowering Team Members", "Influencing Organizational Culture", "Managing Change Effectively", "Cultivating Diversity and Inclusion", "Coaching and Mentoring Others", "Handling Difficult Conversations", "Building High-Performing Teams", "Setting Clear Expectations", "Providing Constructive Feedback", "Adaptive Leadership", "Ethical Leadership Practices", "Crisis Management and Leadership", "Resilient Leadership", "Leading Remote Teams", "Leading Through Uncertainty", "Servant Leadership"}},
				{Name: "Motivation and Goals", Items: []string{"Setting SMART Goals", "Finding Your Purpose", "Staying Motivated in Difficult Times", "Overcoming Procrastination", "Cultivating Self-Discipline", "Harnessing Intrinsic Motivation", "Visualizing Success", "Creating Action Plans", "Breaking Goals into Manageable Steps", "Finding Inspiration in Role Models", "Celebrating Milestones", "Creating a Positive Environment", "Developing a Growth Mindset", "Aligning Goals with Values", "Creating a Vision Board", "Setting Personal and Professional Goals", "Building Confidence and Self-Esteem", "Overcoming Fear of Failure", "Using Affirmations and Mantras", "Finding Supportive Accountability Partners", "Embracing the Power of Habits", "Fostering Resilience in Pursuit of Goals", "Adjusting Goals as Needed", "Setting Priorities", "Reflecting on Progress and Adjusting Strategies"}},
				{Name: "Spiritual Development", Items: []string{"Meditation Practices", "Mindfulness in Daily Life", "Connecting with Nature", "Exploring Sacred Texts", "Deepening Prayer", "Spiritual Growth Journey", "Cultivating Compassion", "Spiritual Community Engagement", "Living Spiritual Values", "Mystical Experiences", "Seeking Spiritual Guidance", "Letting Go and Forgiveness", "Finding Purpose", "Embracing Gratitude", "Aligning with Spiritual Principles", "Cultivating Intuition", "Chanting and Mantras", "Karma and Dharma Understanding", "Inner Wisdom Exploration", "Spiritual Retreats", "Divine Self Connection", "Self-Reflection Practices", "Soulful Relationships", "Sacred Rituals", "Embracing Presence"}},
				{Name: "Creative Self-Expression", Items: []string{"Exploring Different Art Forms", "Finding Your Creative Voice", "Overcoming Creative Blocks", "Cultivating Inspiration", "Experimenting with Mixed Media", "Using Art as Therapy", "Exploring Creative Writing", "Finding Beauty in Everyday Life", "Embracing Imperfection", "Sharing Your Creative Work", "Engaging in Collaborative Projects", "Exploring Digital Art", "Using Art for Social Change", "Capturing Moments through Photography", "Embracing Spontaneity in Art", "Honoring Your Creative Process", "Finding Joy in Creativity", "Exploring Different Mediums", "Creating Art from Found Objects", "Connecting with Other Creatives", "Celebrating Creativity in Children", "Exploring Art History and Culture", "Embracing Abstract Expressionism", "Finding Creative Outlets in Daily Life", "Using Art to Express Emotions"}},
				{Name: "Self-Discipline", Items: []string{"Setting Clear Goals", "Creating Daily Routines", "Developing a Strong Work Ethic", "Resisting Temptations", "Cultivating Willpower", "Establishing Healthy Habits", "Prioritizing Tasks Effectively", "Overcoming Procrastination", "Building Consistency", "Practicing Time Management", "Embracing Delayed Gratification", "Setting Boundaries", "Managing Distractions", "Holding Yourself Accountable", "Fostering Resilience", "Staying Committed to Long-Term Goals", "Learning from Setbacks", "Creating a Supportive Environment", "Monitoring Progress", "Rewarding Yourself for Achievements", "Developing Patience", "Embracing Challenges", "Adopting a Growth Mindset", "Seeking Continuous Improvement", "Celebrating Small Victories"}},
				{Name: "Mental Resilience", Items: []string{"Understanding Resilience", "Coping with Stress", "Building Emotional Strength", "Developing Coping Strategies", "Cultivating Optimism", "Bouncing Back from Adversity", "Nurturing Self-Compassion", "Embracing Change", "Building a Support Network", "Developing Problem-Solving Skills", "Managing Anxiety", "Enhancing Emotional Regulation", "Practicing Mindfulness", "Building Self-Efficacy", "Strengthening Cognitive Flexibility", "Finding Meaning in Challenges", "Cultivating a Positive Mindset", "Building Social Connections", "Overcoming Setbacks", "Learning from Failures", "Resilience in the Face of Trauma", "Building Resilience in Children", "Developing Adaptability", "Building Resilience in the Workplace", "Thriving in Uncertain Times"}},
				{Name: "Work-Life Balance", Items: []string{"Setting Boundaries", "Prioritizing Self-Care", "Effective Time Management", "Flexibility in Work Hours", "Unplugging from Work", "Realistic Expectations", "Leisure Activities", "Quality Time with Loved Ones", "Mindfulness for Stress Reduction", "Rituals for Balance", "Saying No to Overcommitment", "Remote Work Challenges", "Finding Fulfillment", "Social Connections", "Physical Health", "Clear Goals", "Delegating Tasks", "Digital Boundaries", "Supportive Relationships", "Regular Time Off", "Exercise and Activity", "Relaxing Home Environment", "Personal Development Time", "Vacation Planning", "Celebrating Achievements"}},
				{Name: "Building Routines", Items: []string{"Establishing a Morning Routine", "Creating an Evening Routine", "Designing a Productive Work Routine", "Cultivating a Fitness Routine", "Implementing a Healthy Eating Routine", "Setting a Sleep Routine", "Building a Study Routine", "Establishing a Cleaning Routine", "Developing a Meditation Routine", "Creating a Journaling Routine", "Implementing a Creative Routine", "Setting a Budgeting Routine", "Cultivating a Reading Routine", "Establishing a Time-blocking Routine", "Building a Family Routine", "Creating a Social Media Routine", "Implementing a Digital Detox Routine", "Setting a Goal-setting Routine", "Cultivating a Gratitude Routine", "Establishing a Relaxation Routine", "Building a Hobby Routine", "Creating a Mindfulness Routine", "Implementing a Self-Care Routine", "Setting a Learning Routine", "Cultivating a Reflection Routine"}},
				{Name: "Self-Understanding", Items: []string{"Exploring Personal Values", "Understanding Personal Strengths", "Identifying Core Beliefs", "Cultivating Self-Compassion", "Reflecting on Life Experiences", "Embracing Self-Awareness", "Exploring Personality Traits", "Understanding Emotional Triggers", "Practicing Self-Reflection", "Identifying Areas for Growth", "Understanding Personal Motivations", "Exploring Life Purpose", "Recognizing Patterns of Behavior", "Understanding Your Needs", "Exploring Identity and Self-Image", "Navigating Personal Boundaries", "Reflecting on Past Choices", "Cultivating Authenticity", "Embracing Vulnerability", "Identifying Limiting Beliefs", "Exploring Cultural and Social Influences", "Recognizing Your Impact on Others", "Assessing Personal Resilience", "Understanding Personal Biases", "Embracing Change and Growth"}},
				{Name: "Embracing Change", Items: []string{"Understanding the Nature of Change", "Cultivating Adaptability", "Developing a Growth Mindset", "Embracing Uncertainty", "Navigating Life Transitions", "Recognizing Opportunities in Change", "Building Resilience to Change", "Managing Fear of the Unknown", "Embracing Impermanence", "Learning from Past Changes", "Setting Goals for Personal Growth", "Finding Meaning in Change", "Building Support Systems", "Practicing Self-Compassion in Change", "Celebrating Small Wins", "Cultivating Flexibility", "Embracing New Perspectives", "Building Confidence in Change", "Embracing Change as a Catalyst for Growth", "Developing Coping Strategies for Change", "Recognizing Resistance to Change", "Finding Balance Amidst Change", "Letting Go of the Past", "Embracing Change in Relationships", "Committing to Continuous Learning and Growth"}},
				{Name: "Life Values", Items: []string{"Identifying Core Values", "Prioritizing What Matters Most", "Living Authentically", "Cultivating Integrity", "Nurturing Meaningful Relationships", "Pursuing Personal Growth and Development", "Embracing Compassion and Empathy", "Fostering Kindness and Generosity", "Honoring Diversity and Inclusion", "Valuing Honesty and Transparency", "Striving for Excellence and Achievement", "Embracing Creativity and Innovation", "Respecting Individuality and Autonomy", "Cultivating Resilience and Perseverance", "Nurturing a Sense of Purpose and Meaning", "Fostering Gratitude and Appreciation", "Embracing Balance and Harmony", "Valuing Health and Well-being", "Prioritizing Family and Community", "Seeking Adventure and Exploration", "Embracing Sustainability and Environmental Responsibility", "Valuing Education and Lifelong Learning", "Respecting Cultural and Social Diversity", "Promoting Justice and Equity", "Embracing Spirituality and Transcendence"}},
				{Name: "Psychological Support in Crisis", Items: []string{"Crisis Intervention Techniques", "Emotional Regulation Strategies", "Grounding Techniques for Anxiety", "Coping with Overwhelming Emotions", "Creating a Crisis Safety Plan", "Building Resilience in Crisis", "Self-Compassion Practices", "Identifying Supportive Resources", "Navigating Trauma Responses", "Developing Coping Skills", "Addressing Suicidal Thoughts", "Managing Panic Attacks", "Cultivating Social Support Networks", "Engaging in Relaxation Techniques", "Finding Meaning in Crisis", "Practicing Mindfulness in Difficult Moments", "Balancing Self-Care Activities", "Seeking Immediate Help in Emergencies", "Supporting Loved Ones in Crisis", "Recognizing Signs of Psychological Distress", "Building Hope and Optimism", "Celebrating Progress in Recovery"}},
				{Name: "Social Adaptation", Items: []string{"Building Social Skills", "Navigating Social Situations", "Developing Communication Skills", "Building Confidence in Social Settings", "Practicing Active Listening", "Understanding Social Cues and Norms", "Cultivating Empathy and Understanding", "Building Relationships and Friendships", "Navigating Group Dynamics", "Building Trust in Relationships", "Managing Social Anxiety", "Assertiveness Training", "Setting Boundaries in Social Interactions", "Nurturing Supportive Friendships", "Respecting Diversity and Inclusion", "Resolving Conflicts Peacefully", "Networking for Professional and Personal Growth", "Building Rapport with Others", "Balancing Social and Personal Life", "Building Resilience to Rejection", "Seeking Support from Social Networks", "Embracing New Cultural and Social Environments", "Building Social Confidence", "Engaging in Social Activities and Hobbies", "Celebrating Personal Growth in Social Adaptation"}},
				{Name: "Dealing with Criticism", Items: []string{"Understanding Different Types of Criticism", "Keeping an Open Mind", "Separating Emotions from Feedback", "Listening Carefully", "Seeking Constructive Feedback", "Responding Calmly", "Asking for Clarification", "Finding Value in Feedback", "Avoiding Defensiveness", "Learning from Criticism", "Setting Boundaries", "Choosing Which Criticism to Accept", "Recognizing Biases", "Seeking Supportive Input", "Maintaining Perspective", "Responding Assertively", "Practicing Self-Compassion", "Taking Action on Valid Points", "Letting Go of Unhelpful Critique", "Reflecting on Patterns", "Using Feedback for Improvement", "Embracing Growth Opportunities", "Fostering a Positive Feedback Culture", "Celebrating Progress", "Moving Forward with Confidence"}},
			},
		},
		{
			MainGroup: "Mindful Practices",
			SubGroups: []struct {
				Name  string
				Items []string
			}{
				{Name: "Stress management", Items: []string{"Mindful breathing", "Progressive muscle relaxation", "Stressor journaling", "Mindfulness meditation", "Grounding practice", "Visualization of calm", "Focus redirection", "Cognitive restructuring strategies", "Emotional identification", "Setting personal boundaries", "Worst-case scenario analysis", "Affirmation practice", "Reframing stressful situations", "Short-term isolation (for recovery)", "Psychological distancing (metaphorical use)", "4-7-8 breathing technique", "Self-massage (relaxation)", "Practice of small pleasures (mindfulness of the present)", "Visualization of an inner sanctuary"}},
				{Name: "Reducing depression", Items: []string{"Cognitive reassessment", "Behavioral activation", "Advanced visualization technique", "Gratitude journaling", "Progressive Muscle Relaxation", "Mindfulness exercises", "Therapeutic writing", "Self-hypnosis with affirmations", "Cognitive Behavioral Model Feedback", "Art therapy", "Adaptive skills training", "Maintaining a daily regimen program", "Continuous learning and development", "Cognitive Reconstruction", "Resource Activity Planning"}},
				{Name: "Emotional regulation", Items: []string{"Keeping a diary of emotions", "OBD technique", "Thought Stopping", "Mood Affirmations", "Breath regulation", "Relaxation meditations", "Attention Shifting", "Progressive Muscle Relaxation", "Emotional Intelligence Exercises", "Individual yoga", "CBT techniques", "Self-care methods", "Physical exercises", "Emotion literature", "Letting go of negative thoughts"}},
				{Name: "Better sleep", Items: []string{"Regular sleep mode", "Limit screen time", "Practice relaxation", "Warm bath before bedtime", "Listening to soothing audio", "Providing darkness", "Minimize noise", "Breathing techniques", "Meal regimen", "Daily exercise", "Cognitive therapy", "Use of white noise", "Yoga for relaxation", "Keeping a sleep diary", "Visual thinking"}},
				{Name: "Anxiety relief", Items: []string{"Breathing exercises", "Mindfulness practice", "Progressive muscle relaxation", "Physical exercises", "Meditation", "Earth-Recognition Technique", "Time Management", "Self-massage", "Keeping an anxiety diary", "Cognitive reassessment", "Art therapy", "Social support", "Reflection of automatic thoughts", "Grounding exercises", "Psychological counseling", "Caffeine avoidance", "Hand/arm massage", "Gratitude Technique", "Walking under rapture", "Mimic relaxation"}},
				{Name: "Working with negative thoughts", Items: []string{"Cognitive reassessment", "Keeping a thought diary", "Identifying automatic thoughts", "Analyzing the pros and cons of facts", "Conversion to alternative cognitions", "Hot Chair Technique", "Reforming maladaptive thoughts", "Thought Stopping Technique", "Communication with a psychotherapist", "Using affirmations", "Meditation on awareness of thoughts", "Socratic dialog method", "Paraphrasing for positivity", "Conscious reasoning", "Using humor techniques", "Working through mental filters", "Writing down \"clues\" and reasoning", "The \"proof and objectivity\" method", "Reality testing experiments", "Visualizing alternative outcomes"}},
				{Name: "Burnout recovery", Items: []string{"Identifying stress triggers", "Setting work boundaries", "Developing a work-rest balance", "Reducing workload", "Practicing mindfulness", "Getting professional support", "Developing a social circle", "Keeping a self-care diary", "Planning a vacation", "Adapting your day to your body's rhythms", "Activities that bring joy", "Creating non-work rituals", "Regular exercise", "Changing work environment", "Reducing demands on yourself"}},
				{Name: "Social skills and communication", Items: []string{"Active listening", "Emotional expression skills", "Practice communicating with new people", "Confident questioning", "Non-verbal communication skills", "Practicing applications in conflict", "Empathy", "Developing small talk skills", "The Art of Complimentation", "Open and receptive posture", "Adapting to group dynamics", "Developing assertive behavior", "Speaking concisely", "Maintaining eye contact", "Reading nonverbal cues", "Pausing before responding", "Adaptive communication in mixed-age groups", "Showing acceptance", "Managing silence", "Clarifying questioning techniques"}},
				{Name: "Anger management", Items: []string{"Deep breathing", "Keeping a diary of emotions", "Art therapy", "Mindfulness practice", "Attention shifting", "Reforming negative thoughts", "Physical exercises", "Active listening", "Time-out techniques", "Relaxation meditation", "Creating a list of triggers", "Developing problem-solving strategies", "Body Calming Exercises", "Practicing affirmations", "Yoga classes", "Cognitive reassessment", "Generating solutions", "Evaluating Consequences", "The Art of Forgiveness", "Meaningful distractions"}},
				{Name: "Effective time management", Items: []string{"Creating a task list", "Prioritizing things to do", "The Tomato Method", "Setting SMART goals", "Delegating tasks", "Time blocking techniques", "End of Day Reflection", "Reducing multitasking", "Breaking down large tasks into steps", "Process automation", "Regular breaks", "Controlling distractions", "Weekly planning", "Gantt chart", "Allocate time for personal development", "Keep a time diary", "Limit the length of meetings", "Balancing work and leisure", "Adopting a \"finished\" time management system", "Analyzing the effectiveness of routines"}},
				{Name: "Breathing practices", Items: []string{"Deep diaphragmatic breathing", "Method 4-7-8", "Alternating nostril breathing", "Square breathing", "Conscious breathing", "Kapalabhati technique", "Breathing through compressed lips", "Buteyko method", "Practice of \"ship\" breathing", "Pulsating breathing", "Breathing with visualization", "Progressive expansion of the breath", "Ujjayi method", "Full Yogic Breathing Technique", "Breathing meditation", "Pranayama", "Reverse Breathing", "Use of rhythmic visualizations", "Mastered relaxation with exhalation", "Breathing with mantra chanting"}},
				{Name: "Developing empathy", Items: []string{"Active listening", "Mindfulness practice", "Reading fiction", "Appreciating and expressing other people's emotions", "Open attention to nonverbal cues  ", "Empathy through dialogues", "Meditation on benevolence", "Maintaining eye contact", "Practicing traces of developing advice", "Listening with open questions", "Discussing events through others' experiences", "Developing self-awareness", "Storytelling techniques", "Rapport exercises", "Reporting the inner conversation", "Practicing constructive dialog", "Awareness of one's own behavior"}},
				{Name: "Developing positive thinking", Items: []string{"Keeping a gratitude journal", "Practicing affirmations", "Re-wording to avoid negativity", "Visualizing success", "Focusing on inner abilities", "Focus on solutions", "Studying inspirational literature", "Highlighting small victories", "Limiting negative environments", "Self-esteem through action", "Memorizing the day's helpful events", "Humor as a tool", "Developing cognitive reflection", "Meditation on positivity", "Creating a visualization board", "Maintaining connections and relationships", "Engaging in dialog through positivity", "Reflecting on successes", "Showing kindness on a regular basis", "Creating a plan for a wish"}},
				{Name: "Personal development", Items: []string{"Practice self-reflection", "Identifying personal values", "Setting life goals", "Reflection on past events", "Creating a mental map", "Studying biographies of inspirational people", "Self-learning through reading", "Developing active listening skills", "Creating a visualization board", "Self-reflective writing", "Visualizing future achievements", "Emotional control training", "Re-evaluation of past experiences", "Socratic dialog methods", "Reflection through creativity", "Personal development planning", "Working with the inner critic"}},
				{Name: "Unknown", Items: []string{"The Tomato Technique", "Making task lists", "Setting micro-tasks", "Time management", "Clearly defining goals", "Ability to delegate", "Creating a work schedule", "Two-minute strategy", "Eliminating distractions", "Prioritizing", "Planning in reverse order", "Using SMART methodology", "Regular breaks", "Detailed planning of the day", "Stimulating motivation", "Progress monitoring", "Use of reminders", "Time control points", "Diary method", "Task demo"}},
				{Name: "Fighting Procrastination", Items: []string{"The \"Do It Now\" method", "Setting clear goals", "Creating a schedule", "Dividing tasks into subtasks", "Identifying the causes of procrastination", "External support", "Managing distractions", "Rewards for completed tasks", "The 5 Minute Technique", "Recording progress", "Creating a positive environment", "Visualization technique", "Using deadlines", "Working in a community of like-minded people", "Developing habits", "Positive self-reflection", "Refusal to be perfect", "Focusing on one task", "Reviewing successes at the end of the day"}},
				{Name: "Strategies for problem solving", Items: []string{"Problem definition", "Information gathering", "Brainstorming", "Analyzing advantages and disadvantages", "Prioritization of solutions", "Setting SMART goals", "Creating a step-by-step plan", "Finding feedback", "Modeling consequences", "Breaking down the problem", "Focus on solutions", "Considering risks", "Using mind maps", "Asking for help", "Evaluating success stories", "Developing creative solutions", "Iterative testing", "Periodic progress checks", "Evaluating and learning from experience", "Implementation of lessons learned"}},
				{Name: "Conflict Management", Items: []string{"Active listening", "Empathy for both sides", "Identifying common interests", "Clarification of facts", "Discussing behind closed doors", "Controlling emotions", "Applying self-messaging techniques", "Finding compromises", "Focusing on problem solving", "Avoiding blame", "Setting clear expectations", "Mirroring technique", "Asking for a mediator", "Temporary \"time-outs\"", "Dealing with mistakes", "Using nonverbal communication techniques", "Developing an action plan", "Showing respect", "Stabilizing the atmosphere", "Periodic feedback"}},
				{Name: "Work-life balance", Items: []string{"Setting boundaries", "Creating a clear schedule", "Prioritize family time", "Organized space for work", "Planning for quality time off", "Regular periods of \"unplugging\"", "Incorporating exercise", "Delegating tasks", "Practicing one-on-one time", "Focus on quality interactions", "Flexible schedules", "Providing \"off\" areas", "Differentiation of goals", "Meditation or relaxation time blocks", "Implementing self-care habits", "Recognizing \"overtime\" time", "Family traditions", "Allocating time for hobbies", "Spending time in nature", "Workload management"}},
				{Name: "Adaptation to change", Items: []string{"Openness to new experiences", "Developing flexibility of thinking", "Setting realistic expectations", "Identifying your resources", "Seeking support", "Positive self-reflection", "Planning for incremental change", "Focusing on learning", "Learning from success stories", "Building stress resilience", "Developing alternative plans", "Allocating time for adaptation", "Analyzing potential prospects", "Using a \"small steps\" strategy", "Identifying positive aspects of change", "Using mindfulness techniques", "Identifying scalable tasks", "Regular evaluation of progress", "Supporting the social circle", "Introducing a new routine"}},
				{Name: "Developing and building new habits", Items: []string{"Setting clear goals", "Creating a concrete action plan", "Selecting triggers for the habit", "Gradually incorporating changes", "Using daily reminders", "Tracking progress", "Rewards for achievements", "Identifying obstacles", "Maintaining flexibility and adaptation", "Aligning new habits with old ones", "Finding motivation and inspiration", "Sharing experiences with another person", "Keeping an improvement journal", "Working through cognitive biases", "Gradual increase in difficulty", "Evaluating long-term results", "Positive thought alteration", "Creating a maintenance plan", "Rationalizing goals", "Taking care of yourself"}},
				{Name: "Unknown 2", Items: []string{"Setting clear goals", "Creating a concrete action plan", "Selecting triggers for the habit", "Gradually incorporating changes", "Using daily reminders", "Tracking progress", "Rewards for achievements", "Identifying obstacles", "Maintaining flexibility and adaptation", "Aligning new habits with old ones", "Finding motivation and inspiration", "Sharing experiences with another person", "Keeping an improvement journal", "Working through cognitive biases", "Gradual increase in difficulty", "Evaluating long-term results", "Positive thought alteration", "Creating a maintenance plan", "Rationalizing goals", "Taking care of yourself"}},
				{Name: "Development of financial literacy", Items: []string{"Budget management training", "Understanding financial terms", "Analyzing and accounting for expenses", "Setting financial goals", "Investment Strategies", "Tax Accounting Basics", "Retirement planning", "Debt Management", "Saving and Savings", "Calculating your credit score", "Using financial apps", "Financial courses and books", "Understanding insurance", "Building a financial reserve", "Strategic planning for purchases", "Financial counseling", "Healthy relationship with money", "Risk assessment", "Separation of needs and wants", "Psychological aspect of finances"}},
				{Name: "Developing emotional intelligence", Items: []string{"Awareness of one's own emotions", "Empathy and sympathy", "Self-regulation of emotional reactions", "Analyzing and controlling triggers", "Active listening", "Stress Management", "Anger Management", "Constructive Expression of Emotions", "Effective Communication", "Emotional support for others", "Conflict Resolution", "Practicing gratitude", "Affirmations and positive thinking", "Reflection of emotional experiences", "Learning to detect nonverbal cues", "Learning to accept criticism", "Reading fiction to understand empathy", "Mindfulness meditation", "Social interaction and difficulties in interpersonal relationships", "Emotional development workshops and trainings"}},
				{Name: "Preparing for Conflict Conversations", Items: []string{"Active listening", "Determining the goals of the conversation", "De-escalation techniques", "Developing assertiveness", "Recognizing and expressing emotions", "Setting clear boundaries", "Using \"self-talk\"", "Empathy and understanding perspectives", "Developing problem-solving strategies", "Respectful attitudes", "Social support", "Time out in an emotional situation", "Recognizing mistakes", "Preparing arguments", "Psychological role of a difficult situation", "Meditation and breathing techniques", "Culture of corporate communication", "Recognizing non-verbal cues"}},
			},
		},
	}

	// Insert default prompts
	for _, p := range defaultPrompts {
		p.MainGroupID = nil // Explicitly set to NULL
		p.SubGroupID = nil  // Explicitly set to NULL
		if err := db.FirstOrCreate(&p, model.Prompt{Name: p.Name}).Error; err != nil {
			return err
		}
	}

	// Step 1: Insert Main Groups
	mainGroupMap := make(map[string]uint) // Cache to store MainGroup IDs
	for _, group := range groups {
		mainGroup := model.PromptMainGroup{Name: group.MainGroup}
		if err := db.FirstOrCreate(&mainGroup, model.PromptMainGroup{Name: group.MainGroup}).Error; err != nil {
			return err
		}
		mainGroupMap[group.MainGroup] = mainGroup.ID
	}

	// Step 2: Insert SubGroups
	subGroupMap := make(map[string]uint) // Cache to store SubGroup IDs
	for _, group := range groups {
		mainGroupID := mainGroupMap[group.MainGroup]
		basePromptID := uint(1)
		for _, subGroup := range group.SubGroups {
			subGroupRecord := model.PromptSubGroup{
				Name:         subGroup.Name,
				MainGroupID:  mainGroupID,
				BasePromptID: &basePromptID,
			}
			if err := db.FirstOrCreate(&subGroupRecord, model.PromptSubGroup{Name: subGroup.Name, MainGroupID: mainGroupID}).Error; err != nil {
				return err
			}
			subGroupMap[subGroup.Name] = subGroupRecord.ID
		}
	}

	// Step 3: Insert Prompts
	for _, group := range groups {
		for _, subGroup := range group.SubGroups {
			subGroupID := subGroupMap[subGroup.Name]
			mainGroupID := mainGroupMap[group.MainGroup]
			for _, item := range subGroup.Items {
				prompt := model.Prompt{
					Name:        item,
					Content:     "Generated content for " + item, // Placeholder content
					Temperature: 1.05,                            // Default temperature
					MaxTokens:   1000,                            // Default max tokens
					ModelName:   "GPT-4o",
					MainGroupID: &mainGroupID, // Use pointer to mainGroupID
					SubGroupID:  &subGroupID,  // Use pointer to subGroupID
				}
				if err := db.FirstOrCreate(&prompt, model.Prompt{
					Name:        item,
					MainGroupID: &mainGroupID, // Use pointer for comparison
					SubGroupID:  &subGroupID,  // Use pointer for comparison
				}).Error; err != nil {
					return err
				}
			}
		}
	}

	// Step 4: Insert Default Non-Grouped Prompts
	for _, p := range defaultPrompts {
		if err := db.FirstOrCreate(&p, model.Prompt{Name: p.Name}).Error; err != nil {
			return err
		}
	}

	// Step 5: Custom update for PromptSubGroup
	if err := db.Model(&model.PromptSubGroup{}).
		Where("base_prompt_id IS NULL").
		Update("base_prompt_id", 1).Error; err != nil {
		return fmt.Errorf("failed to update BasePrompt for subgroups: %w", err)
	}

	log.Println("Custom migrations applied successfully.")
	return nil
}
