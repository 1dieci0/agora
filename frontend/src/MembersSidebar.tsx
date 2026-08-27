
import type { Member } from "./types";

type MembersSidebarProps = {
  members: Member[];
};

function MembersSidebar({
  members,
}: MembersSidebarProps) {
  const owners = members.filter(
    (member) => member.role === "owner",
  );

  const admins = members.filter(
    (member) => member.role === "admin",
  );

  const regularMembers = members.filter(
    (member) => member.role === "member",
  );

  return (
    <aside className="members-sidebar">
      {owners.length > 0 && (
        <section className="member-group">
          <h3>OWNER — {owners.length}</h3>

          {owners.map((member) => (
            <div
              className="member"
              key={member.id}
            >
              <span className="member-avatar">
                {member.username
                  .charAt(0)
                  .toUpperCase()}
              </span>

              <span className="member-name">
                {member.username}
              </span>
            </div>
          ))}
        </section>
      )}

      {admins.length > 0 && (
        <section className="member-group">
          <h3>ADMINS — {admins.length}</h3>

          {admins.map((member) => (
            <div
              className="member"
              key={member.id}
            >
              <span className="member-avatar">
                {member.username
                  .charAt(0)
                  .toUpperCase()}
              </span>

              <span className="member-name">
                {member.username}
              </span>
            </div>
          ))}
        </section>
      )}

      {regularMembers.length > 0 && (
        <section className="member-group">
          <h3>MEMBERS — {regularMembers.length}</h3>

          {regularMembers.map((member) => (
            <div
              className="member"
              key={member.id}
            >
              <span className="member-avatar">
                {member.username
                  .charAt(0)
                  .toUpperCase()}
              </span>

              <span className="member-name">
                {member.username}
              </span>
            </div>
          ))}
        </section>
      )}
    </aside>
  );
}

export default MembersSidebar;
